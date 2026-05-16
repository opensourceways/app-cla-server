#!/bin/bash
# Unified Go Coverage Check Script
# Used across all opensourceways projects
# Full coverage threshold: configurable (default 10%)
# Incremental coverage threshold: configurable (default 80%)

set -e

# Configuration (from environment or defaults)
FULL_THRESHOLD=${FULL_COVERAGE_THRESHOLD:-10.0}
INC_THRESHOLD=${INC_COVERAGE_THRESHOLD:-80.0}
COVERAGE_FILE=${COVERAGE_FILE:-coverage.out}
TEST_PACKAGES=${TEST_PACKAGES:-./...}
GO_TEST_FLAGS=${GO_TEST_FLAGS:--short}

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Result status
GLOBAL_SUCCESS=true

# Log functions
log_info() { echo "[INFO] $1"; }
log_success() { echo "${GREEN}[PASS]${NC} $1"; }
log_fail() { echo "${RED}[FAIL]${NC} $1"; }
log_warn() { echo "${YELLOW}[WARN]${NC} $1"; }

# Run tests and generate coverage
run_coverage_test() {
    log_info "Running tests with coverage on: $TEST_PACKAGES"
    
    # Set GOPROXY (respects existing environment)
    export GOPROXY=${GOPROXY:-https://proxy.golang.org,direct}
    
    # Run tests with proper error handling
    if ! go test -v -coverprofile="$COVERAGE_FILE" $GO_TEST_FLAGS "$TEST_PACKAGES"; then
        log_fail "Go test execution failed"
        return 1
    fi
    
    if [ ! -f "$COVERAGE_FILE" ]; then
        log_fail "No coverage file generated"
        return 1
    fi
    
    log_info "Coverage file generated: $COVERAGE_FILE"
    return 0
}

# Calculate full coverage
calculate_full_coverage() {
    local coverage_file=$1
    
    if [ ! -f "$coverage_file" ]; then
        log_fail "Coverage file not found: $coverage_file"
        return 1
    fi
    
    # Extract total coverage percentage
    coverage=$(go tool cover -func="$coverage_file" | grep total | awk '{print $3}' | sed 's/%//')
    full_cov="$coverage"
    
    if [ -z "$coverage" ]; then
        log_fail "Unable to parse full coverage"
        return 1
    fi
    
    echo ">>> Full Coverage: ${coverage}%"
    
    # Check threshold (use awk -v for safe variable passing)
    if awk -v cov="$coverage" -v th="$FULL_THRESHOLD" 'BEGIN { exit !(cov < th) }'; then
        log_fail "Full coverage below threshold (${coverage}% < ${FULL_THRESHOLD}%)"
        GLOBAL_SUCCESS=false
        return 1
    else
        log_success "Full coverage meets threshold (${coverage}% >= ${FULL_THRESHOLD}%)"
        return 0
    fi
}

# Calculate incremental coverage (based on Statement Block)
calculate_incremental_coverage() {
    local coverage_file=$1
    local diff_file=$2
    
    echo ">>> Incremental Coverage Analysis (Block-based)"
    
    # Check diff file
    if [ ! -f "$diff_file" ]; then
        log_warn "No diff file found, skipping incremental coverage check"
        return 0
    fi
    
    # Extract changed lines (excluding test files)
    local changed_lines=$(awk '
        # Match .go files, exclude _test.go
        /^\+\+\+ b\/.*\.go/ && !/.*_test\.go/ {
            file = substr($NF, 3);
            active=1;
            next
        }
        # Close on new file header
        /^\+\+\+ b\// {active=0; next}
        
        # Parse hunk header for line numbers
        active == 1 && /^@@/ {
            split($3, a, ",");
            line = substr(a[1], 2);
            next
        }
        
        # Filter: only executable code lines
        active == 1 && /^\+/ && !/^\+\+\+/ {
            content = substr($0, 2);
            gsub(/[ \t\r\n]/, "", content);
            
            # Exclude: empty, braces, package/import, comments, type definitions
            if (content == "" ||
                content ~ /^}/ ||
                content ~ /^{/ ||
                content ~ /^\)/ ||
                content ~ /^package/ ||
                content ~ /^import/ ||
                content ~ /^\/\// ||
                content ~ /^\/\*/ ||
                content ~ /^type.*struct{/ ||
                content ~ /^type.*interface{/) {
                line++;
                next;
            }
            
            print file ":" line;
            line++;
            next;
        }
        
        # Maintain line count
        active == 1 && !/^\-/ { line++ }
    ' "$diff_file")
    
    if [ -z "$changed_lines" ]; then
        log_warn "No executable code changes, skipping incremental coverage"
        return 0
    fi
    
    echo ">>> Changed lines found:"
    echo "$changed_lines" | head -20
    
    # Use associative array for processed blocks (more reliable than string matching)
    declare -A processed_blocks=()
    
    # Read coverage file into memory once (O(N+M) instead of O(N*M))
    declare -A coverage_blocks_by_file
    while IFS= read -r cover_line; do
        [[ -z "$cover_line" ]] && continue
        # Skip coverage file header line (mode: set)
        [[ "$cover_line" == mode:* ]] && continue
        local file_path="${cover_line%%:*}"
        local data="${cover_line##*:}"
        coverage_blocks_by_file[$file_path]+="${data}\n"
    done < "$coverage_file"
    
    local total_inc_stmts=0
    local covered_inc_stmts=0
    
    for info in $changed_lines; do
        local target_file="${info%:*}"
        local target_line="${info#*:}"
        
        # Get blocks for this file from memory
        local blocks_data="${coverage_blocks_by_file[$target_file]}"
        if [ -z "$blocks_data" ]; then
            continue
        fi
        
        # Find blocks containing this line
        while IFS= read -r block_data; do
            [[ -z "$block_data" ]] && continue
            
            # Parse: start_line.col,end_line.col stmts hits
            local line_range stmts hits
            read -r line_range stmts hits <<< "$block_data"
            
            local start_line="${line_range%%.*}"
            local tmp="${line_range#*,}"
            local end_line="${tmp%%.*}"
            
            # Check if changed line is within block range
            if ((target_line >= start_line && target_line <= end_line)); then
                local block_uid="${target_file}:${line_range}"
                
                if [[ -z "${processed_blocks[$block_uid]}" ]]; then
                    total_inc_stmts=$((total_inc_stmts + stmts))
                    
                    if ((hits > 0)); then
                        covered_inc_stmts=$((covered_inc_stmts + stmts))
                        echo "  ✅ Covered: ${block_uid} (${stmts} stmts)"
                    else
                        echo "  ❌ Uncovered: ${block_uid} (${stmts} stmts)"
                    fi
                    
                    processed_blocks[$block_uid]=1
                fi
                break
            fi
        done <<< "$(echo -e "$blocks_data")"
    done
    
    # Calculate result
    echo "--------------------------------------"
    echo "Incremental Coverage Report (Block-based):"
    
    if [ "$total_inc_stmts" -eq 0 ]; then
        inc_cov="100.0"
        local status="PASS"
        log_warn "No incremental statements to cover, default pass"
    else
        local result=$(awk -v covered="$covered_inc_stmts" -v total="$total_inc_stmts" -v th="$INC_THRESHOLD" 'BEGIN {
            rate = (covered / total) * 100;
            printf "%.1f|%s", rate, (rate < th ? "FAIL" : "PASS");
        }')
        
        inc_cov="${result%|*}"
        local status="${result#*|}"
        
        echo "Total statements: $total_inc_stmts"
        echo "Covered statements: $covered_inc_stmts"
        echo "Incremental coverage: ${inc_cov}%"
    fi
    
    echo "--------------------------------------"
    
    if [ "$status" = "FAIL" ]; then
        log_fail "Incremental coverage below threshold (${inc_cov}% < ${INC_THRESHOLD}%)"
        GLOBAL_SUCCESS=false
        return 1
    else
        log_success "Incremental coverage meets threshold (${inc_cov}% >= ${INC_THRESHOLD}%)"
        return 0
    fi
}

# Generate PR diff file
generate_diff_file() {
    local output_file=$1
    
    if [ -f .git/shallow ]; then
        log_fail "Shallow clone detected! This workflow requires full git history (fetch-depth: 0)"
        log_fail "Incremental coverage analysis cannot work with shallow clones"
        log_fail "Please ensure checkout action uses fetch-depth: 0"
        return 1
    fi
    
    # GitHub Actions environment
    if [ -n "$GITHUB_EVENT_PATH" ]; then
        local base_ref="${GITHUB_BASE_REF:-main}"
        local head_ref="${GITHUB_HEAD_REF:-$(git rev-parse HEAD)}"
        
        log_info "GitHub Actions: base=$base_ref, head=$head_ref"
        
        git fetch origin "$base_ref" 2>/dev/null || true
        if ! git diff "origin/$base_ref"..."$head_ref" > "$output_file" 2>/dev/null; then
            log_fail "Failed to generate diff. Check fetch-depth."
            return 1
        fi
    else
        # Local environment
        log_info "Local: using git diff HEAD~1"
        git diff HEAD~1 > "$output_file" 2>/dev/null || git diff HEAD > "$output_file"
    fi
    
    if [ ! -s "$output_file" ]; then
        log_warn "Generated diff file is empty or missing"
        return 1
    fi
    
    log_info "Diff file generated: $output_file"
    return 0
}

# Print final report
print_final_report() {
    local full_cov=$1
    local inc_cov=$2
    
    echo ""
    echo "==========================================================="
    echo "              GO COVERAGE CHECK FINAL REPORT               "
    echo "==========================================================="
    printf "| %-35s | %-12s | %-8s |\n" "TYPE" "COVERAGE" "RESULT"
    echo "-----------------------------------------------------------"
    
    if [ "$GLOBAL_SUCCESS" = true ]; then
        printf "| %-35s | %-12s | %-8s |\n" "Full Coverage" "${full_cov:-N/A}%" "PASS"
        printf "| %-35s | %-12s | %-8s |\n" "Incremental Coverage" "${inc_cov:-N/A}%" "PASS"
        echo "==========================================================="
        log_success "All coverage checks passed"
        return 0
    else
        printf "| %-35s | %-12s | %-8s |\n" "Full Coverage" "${full_cov:-N/A}%" "FAIL"
        printf "| %-35s | %-12s | %-8s |\n" "Incremental Coverage" "${inc_cov:-N/A}%" "FAIL"
        echo "==========================================================="
        log_fail "Coverage checks failed!"
        return 1
    fi
}

# Main function
main() {
    local diff_file="pr_changes.patch"
    
    echo ">>> Go Coverage Check"
    echo ">>> Full coverage threshold: ${FULL_THRESHOLD}%"
    echo ">>> Incremental coverage threshold: ${INC_THRESHOLD}%"
    echo ""
    
    # 1. Run tests and generate coverage
    if ! run_coverage_test; then
        exit 1
    fi
    
    # 2. Calculate full coverage
    echo ""
    calculate_full_coverage "$COVERAGE_FILE"
    
    # 3. Generate diff and calculate incremental coverage
    echo ""
    if generate_diff_file "$diff_file"; then
        calculate_incremental_coverage "$COVERAGE_FILE" "$diff_file"
    fi
    local inc_cov=${inc_cov:-N/A}
    
    # 4. Print final report
    echo ""
    if print_final_report "$full_cov" "$inc_cov"; then
        exit 0
    else
        exit 1
    fi
}

main