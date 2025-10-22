# VHS tape for quick-start local testing demo
#
# This is a condensed version showing just the essential commands
# to get started with local testing.
#
# Usage: vhs demo-quick-start.tape
# Output: demo-quick-start.gif

# Terminal setup
Output demo-quick-start.gif
Set Shell "bash"
Set FontSize 16
Set Width 1000
Set Height 600
Set Padding 15
Set Theme "Dracula"

# Navigate to test directory
Type "cd .github/test"
Enter
Sleep 1s

Type "# Quick Start: Local Testing with Docker"
Sleep 500ms
Enter
Sleep 1s

# Step 1: Start services
Type "# Step 1: Start Docker services"
Sleep 500ms
Enter
Type "docker compose up --wait"
Enter
Sleep 6s

# Step 2: Verify services
Type "# Step 2: Verify services are running"
Sleep 500ms
Enter
Type "docker compose ps"
Enter
Sleep 2s

# Step 3: Run tests
Type "# Step 3: Run the test script"
Sleep 500ms
Enter
Type "../scripts/run-e2e-test.sh"
Enter
Sleep 6s

Type "# To stop services: docker compose down"
Sleep 500ms
Enter
Sleep 2s
