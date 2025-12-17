# Library: print.sh
# Description: Colorized output and status message functions
# Usage: source ./scripts/print.sh

# Color definitions
# Check if terminal supports colors
if [ -t 1 ] && command -v tput > /dev/null 2>&1; then
    PRINT_COLOR_RED="$(tput setaf 1)"
    PRINT_COLOR_GREEN="$(tput setaf 2)"
    PRINT_COLOR_YELLOW="$(tput setaf 3)"
    PRINT_COLOR_BLUE="$(tput setaf 4)"
    PRINT_COLOR_MAGENTA="$(tput setaf 5)"
    PRINT_COLOR_CYAN="$(tput setaf 6)"
    PRINT_COLOR_WHITE="$(tput setaf 7)"
    PRINT_COLOR_BOLD="$(tput bold)"
    PRINT_COLOR_RESET="$(tput sgr0)"
else
    PRINT_COLOR_RED=""
    PRINT_COLOR_GREEN=""
    PRINT_COLOR_YELLOW=""
    PRINT_COLOR_BLUE=""
    PRINT_COLOR_MAGENTA=""
    PRINT_COLOR_CYAN=""
    PRINT_COLOR_WHITE=""
    PRINT_COLOR_BOLD=""
    PRINT_COLOR_RESET=""
fi

# Namespace: print
# All functions in this library use the print:: prefix

print::header() {
    local message="${1}"
    printf "\n%s%s=== %s ===%s\n\n" "${PRINT_COLOR_BOLD}" "${PRINT_COLOR_CYAN}" "${message}" "${PRINT_COLOR_RESET}"
}

print::success() {
    local message="${1}"
    printf "%s[SUCCESS]%s %s\n" "${PRINT_COLOR_GREEN}" "${PRINT_COLOR_RESET}" "${message}"
}

print::error() {
    local message="${1}"
    printf "%s[ERROR]%s %s\n" "${PRINT_COLOR_RED}" "${PRINT_COLOR_RESET}" "${message}" >&2
}

print::warning() {
    local message="${1}"
    printf "%s[WARNING]%s %s\n" "${PRINT_COLOR_YELLOW}" "${PRINT_COLOR_RESET}" "${message}"
}

print::info() {
    local message="${1}"
    printf "%s[INFO]%s %s\n" "${PRINT_COLOR_BLUE}" "${PRINT_COLOR_RESET}" "${message}"
}

print::step() {
    local step_number="${1}"
    local message="${2}"
    printf "%s[%s]%s %s\n" "${PRINT_COLOR_MAGENTA}" "${step_number}" "${PRINT_COLOR_RESET}" "${message}"
}

print::debug() {
    local message="${1}"
    if [ "${DEBUG:-false}" = "true" ]; then
        printf "%s[DEBUG]%s %s\n" "${PRINT_COLOR_WHITE}" "${PRINT_COLOR_RESET}" "${message}"
    fi
}
