# VHS tape for demonstrating local testing with Docker and DIMP
#
# This tape shows how to start Docker services and run local tests
# for the Aether pipeline with DIMP integration.
#
# Usage: vhs demo-local-testing.tape
# Output: demo-local-testing.gif

# Set up the terminal
Output demo-local-testing.gif
Set Shell "bash"
Set FontSize 14
Set Width 1200
Set Height 700
Set Padding 10
Set Theme "Dracula"

# Start with a clean state
Type "cd .github/test"
Enter
Sleep 1s

# Show the directory contents
Type "# Let's explore the test environment"
Sleep 500ms
Enter
Type "ls -l"
Enter
Sleep 2s

# Show the README
Type "# First, let's check the README for instructions"
Sleep 500ms
Enter
Type "cat README.md | head -30"
Enter
Sleep 3s

# Start Docker Compose services
Type "# Now let's start the Docker services"
Sleep 500ms
Enter
Type "docker compose up --wait"
Enter
Sleep 8s

# Check service status
Type "# Check if all services are running"
Sleep 500ms
Enter
Type "docker compose ps"
Enter
Sleep 2s

# Get the DIMP port
Type "# Get the dynamically assigned DIMP service port"
Sleep 500ms
Enter
Type "docker compose port fhir-pseudonymizer 8080"
Enter
Sleep 2s

# Run the test script
Type "# Now let's run the test script to verify everything works"
Sleep 500ms
Enter
Type "./test-dimp.sh"
Enter
Sleep 8s

# Show the logs if needed
Type "# The test passed! Let's check the service logs"
Sleep 500ms
Enter
Type "docker compose logs --tail=20 fhir-pseudonymizer"
Enter
Sleep 3s

# Show how to use Make for testing
Type "# We can also use Make for easier testing"
Sleep 500ms
Enter
Type "make help"
Enter
Sleep 2s

# Show available make targets
Type "# Let's see what testing options are available"
Sleep 500ms
Enter
Type "make dimp-test"
Enter
Sleep 5s

# Clean up
Type "# Finally, let's stop the services"
Sleep 500ms
Enter
Type "docker compose down"
Enter
Sleep 3s

# Final message
Type "# Done! Services are stopped and test completed successfully"
Sleep 500ms
Enter
Sleep 2s
