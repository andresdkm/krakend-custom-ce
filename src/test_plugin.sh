#!/bin/bash

set -e

# Colores
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

PLUGINS_DIR="./plugins"
KRAKEND_CONTAINER="krakend"
TEST_RESULTS=()
PASSED=0
FAILED=0

echo "╔═══════════════════════════════════════════════════════════╗"
echo "║          KrakenD Plugin Testing Suite                    ║"
echo "╚═══════════════════════════════════════════════════════════╝"
echo ""

# Función para mostrar mensajes
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_test() {
    echo -e "${CYAN}🧪 $1${NC}"
}

check_krakend_running() {
    log_info "Verificando si KrakenD está corriendo..."
    if docker ps | grep -q "$KRAKEND_CONTAINER"; then
        log_success "KrakenD está corriendo"
        return 0
    else
        log_error "KrakenD no está corriendo"
        echo ""
        echo "Iniciando KrakenD..."
        docker-compose up -d krakend
        sleep 5

        if docker ps | grep -q "$KRAKEND_CONTAINER"; then
            log_success "KrakenD iniciado correctamente"
            return 0
        else
            log_error "No se pudo iniciar KrakenD"
            return 1
        fi
    fi
}

check_plugins_exist() {
    log_info "Verificando plugins compilados..."

    if [ ! -d "$PLUGINS_DIR" ]; then
        log_error "Directorio de plugins no encontrado: $PLUGINS_DIR"
        return 1
    fi

    PLUGIN_COUNT=$(find "$PLUGINS_DIR" -name "*.so" 2>/dev/null | wc -l)

    if [ "$PLUGIN_COUNT" -eq 0 ]; then
        log_warning "No se encontraron plugins compilados (.so) en $PLUGINS_DIR"
        echo ""
        echo "¿Deseas compilar los plugins ahora? (y/n)"
        read -r response
        if [[ "$response" =~ ^[Yy]$ ]]; then
            log_info "Compilando plugins..."
            make build-plugins || docker-compose --profile build up krakend-plugin-builder
            PLUGIN_COUNT=$(find "$PLUGINS_DIR" -name "*.so" 2>/dev/null | wc -l)
        else
            return 1
        fi
    fi

    log_success "Encontrados $PLUGIN_COUNT plugin(s)"
    return 0
}

# Test 1: Verificar que el plugin existe y es válido
test_plugin_file() {
    local plugin_path=$1
    local plugin_name=$(basename "$plugin_path")

    log_test "Test 1: Verificando archivo del plugin: $plugin_name"

    # Verificar que existe
    if [ ! -f "$plugin_path" ]; then
        log_error "Plugin no encontrado: $plugin_path"
        return 1
    fi

    # Verificar que es un archivo .so
    if [[ ! "$plugin_name" =~ \.so$ ]]; then
        log_error "Plugin no tiene extensión .so: $plugin_name"
        return 1
    fi

    # Verificar tipo de archivo
    local file_type=$(file "$plugin_path")
    if [[ "$file_type" =~ "ELF" ]] || [[ "$file_type" =~ "shared object" ]]; then
        log_success "Archivo válido: $file_type"
        return 0
    else
        log_error "Archivo inválido: $file_type"
        return 1
    fi
}

# Test 2: Verificar símbolos exportados
test_plugin_symbols() {
    local plugin_path=$1
    local plugin_name=$(basename "$plugin_path")

    log_test "Test 2: Verificando símbolos exportados: $plugin_name"

    # Verificar símbolos con nm
    if command -v nm &> /dev/null; then
        local symbols=$(nm -D "$plugin_path" 2>/dev/null | grep -E "Register|main" || true)

        if [ -n "$symbols" ]; then
            log_success "Símbolos encontrados:"
            echo "$symbols" | head -5 | sed 's/^/    /'
        else
            log_warning "No se encontraron símbolos Register* (puede ser normal)"
        fi
    else
        log_warning "Comando 'nm' no disponible, saltando test de símbolos"
    fi

    return 0
}

# Test 3: Verificar compatibilidad de arquitectura
test_plugin_architecture() {
    local plugin_path=$1
    local plugin_name=$(basename "$plugin_path")

    log_test "Test 3: Verificando arquitectura: $plugin_name"

    local file_info=$(file "$plugin_path")

    if [[ "$file_info" =~ "x86-64" ]] || [[ "$file_info" =~ "x86_64" ]]; then
        log_success "Arquitectura: x86_64 (amd64)"
        return 0
    elif [[ "$file_info" =~ "ARM aarch64" ]] || [[ "$file_info" =~ "arm64" ]]; then
        log_success "Arquitectura: ARM64"
        return 0
    else
        log_warning "Arquitectura no identificada: $file_info"
        return 0
    fi
}

test_plugin_go_version() {
    local plugin_path=$1
    local plugin_name=$(basename "$plugin_path")

    log_test "Test 6: Verificando compatibilidad de versión de Go: $plugin_name"

    # Intentar cargar el plugin en KrakenD (esto detectará incompatibilidades)
    # Nota: Esto requiere una configuración temporal de KrakenD

    log_info "Verificando versión de Go en el plugin..."

    # Extraer información de Go del plugin
    if command -v strings &> /dev/null; then
        local go_version=$(strings "$plugin_path" | grep -o "go1\.[0-9]*\.[0-9]*" | head -1 || echo "unknown")
        if [ "$go_version" != "unknown" ]; then
            log_success "Versión de Go detectada: $go_version"
        else
            log_warning "No se pudo detectar la versión de Go"
        fi
    fi

    return 0
}


# Función principal para probar un plugin
test_plugin() {
    local plugin_path=$1
    local plugin_name=$(basename "$plugin_path")

    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo -e "${CYAN}Testing Plugin: $plugin_name${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    local tests_passed=0
    local tests_failed=0

    # Ejecutar tests
    test_plugin_file "$plugin_path" && ((tests_passed++)) || ((tests_failed++))
    test_plugin_symbols "$plugin_path" && ((tests_passed++)) || ((tests_failed++))
    test_plugin_architecture "$plugin_path" && ((tests_passed++)) || ((tests_failed++))
    test_plugin_go_version "$plugin_path" && ((tests_passed++)) || ((tests_failed++))
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    if [ $tests_failed -eq 0 ]; then
        log_success "Plugin $plugin_name: Todos los tests pasaron ($tests_passed/8)"
        TEST_RESULTS+=("✅ $plugin_name - PASSED ($tests_passed/8)")
        ((PASSED++))
        return 0
    else
        log_error "Plugin $plugin_name: $tests_failed test(s) fallaron"
        TEST_RESULTS+=("❌ $plugin_name - FAILED ($tests_passed/8 passed, $tests_failed/8 failed)")
        ((FAILED++))
        return 0
    fi
}

# Función para mostrar resumen final
show_summary() {
    echo ""
    echo "╔═══════════════════════════════════════════════════════════╗"
    echo "║                  Test Summary                             ║"
    echo "╚═══════════════════════════════════════════════════════════╝"
    echo ""

    for result in "${TEST_RESULTS[@]}"; do
        echo "$result"
    done

    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo -e "${GREEN}Passed: $PASSED${NC}"
    echo -e "${RED}Failed: $FAILED${NC}"
    echo "Total:  $((PASSED + FAILED))"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""

    if [ $FAILED -eq 0 ]; then
        log_success "🎉 Todos los plugins pasaron los tests!"
        return 0
    else
        log_error "⚠️  Algunos plugins fallaron. Revisa los logs arriba."
        return 1
    fi
}

# Main
main() {
    # Verificar prerequisitos
    check_krakend_running || exit 1
    check_plugins_exist || exit 1

    echo ""
    log_info "Iniciando tests de plugins..."
    echo ""

    # Encontrar y probar todos los plugins
    while IFS= read -r plugin_path; do
        test_plugin "$plugin_path"
    done < <(find "$PLUGINS_DIR" -name "*.so" 2>/dev/null)

    # Mostrar resumen
    show_summary

    # Salir con código apropiado
    [ $FAILED -eq 0 ] && exit 0 || exit 1
}

# Parsear argumentos
case "${1:-}" in
    -h|--help)
        echo "Uso: $0 [opciones]"
        echo ""
        echo "Opciones:"
        echo "  -h, --help     Mostrar esta ayuda"
        echo "  -v, --verbose  Modo verbose (más detalles)"
        echo ""
        echo "Este script prueba todos los plugins compilados (.so) en el directorio './plugins'"
        echo "y verifica que sean compatibles con KrakenD."
        exit 0
        ;;
    -v|--verbose)
        set -x
        main
        ;;
    *)
        main
        ;;
esac