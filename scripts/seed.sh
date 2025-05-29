#!/bin/bash

# Social Media API Data Generation Script
# Usage: ./scripts/seed.sh [options]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
USERS=50
POSTS=5
CLEAN=false
VERBOSE=false
ENV_FILE=".env"

# Function to print colored output
print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_banner() {
    echo -e "${BLUE}"
    cat << "EOF"
╔══════════════════════════════════════════════════════════════╗
║                 SOCIAL MEDIA DATA SEEDER                     ║
║                                                              ║
║  🎲 Generate realistic data for your Social Media API       ║
║  📊 Complete with users, posts, relationships & more!       ║
╚══════════════════════════════════════════════════════════════╝
EOF
    echo -e "${NC}"
}

# Function to show help
show_help() {
    cat << EOF
Social Media API Data Generation Script

USAGE:
    ./scripts/seed.sh [OPTIONS]

OPTIONS:
    -u, --users <number>    Number of users to generate (default: 50)
    -p, --posts <number>    Number of posts per user (default: 5)
    -c, --clean            Clean existing data before generation
    -v, --verbose          Verbose output
    --env <file>           Environment file to use (default: .env)
    --quick                Quick setup (10 users, 3 posts each)
    --demo                 Demo setup (100 users, 8 posts each)
    --large                Large setup (500 users, 10 posts each)
    -h, --help             Show this help message

EXAMPLES:
    # Basic usage
    ./scripts/seed.sh

    # Clean database and generate 100 users with 10 posts each
    ./scripts/seed.sh --clean --users 100 --posts 10

    # Quick demo setup
    ./scripts/seed.sh --quick --clean

    # Large dataset with verbose output
    ./scripts/seed.sh --large --verbose

PRESETS:
    --quick     10 users, 3 posts each (fast)
    --demo      100 users, 8 posts each (demonstration)
    --large     500 users, 10 posts each (performance testing)

ENVIRONMENT:
    Make sure your .env file is configured with proper database settings:
    - DATABASE_URI
    - DATABASE_NAME
    - JWT_SECRET_KEY

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -u|--users)
            USERS="$2"
            shift 2
            ;;
        -p|--posts)
            POSTS="$2"
            shift 2
            ;;
        -c|--clean)
            CLEAN=true
            shift
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        --env)
            ENV_FILE="$2"
            shift 2
            ;;
        --quick)
            USERS=10
            POSTS=3
            shift
            ;;
        --demo)
            USERS=100
            POSTS=8
            shift
            ;;
        --large)
            USERS=500
            POSTS=10
            shift
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Check if we're in the right directory
if [[ ! -f "go.mod" ]]; then
    print_error "Please run this script from the project root directory"
    exit 1
fi

# Check if environment file exists
if [[ ! -f "$ENV_FILE" ]]; then
    print_error "Environment file '$ENV_FILE' not found"
    print_info "Please create a .env file with your database configuration"
    exit 1
fi

print_banner

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed or not in PATH"
    exit 1
fi

# Check if required directories exist
if [[ ! -d "cmd/seed" ]]; then
    print_info "Creating seed command directory..."
    mkdir -p cmd/seed
fi

# Build arguments for the Go command
GO_ARGS=""
if [[ "$CLEAN" == "true" ]]; then
    GO_ARGS="$GO_ARGS --clean"
fi
if [[ "$VERBOSE" == "true" ]]; then
    GO_ARGS="$GO_ARGS --verbose"
fi

GO_ARGS="$GO_ARGS --users $USERS --posts $POSTS"

print_info "Configuration:"
echo "  Users to generate: $USERS"
echo "  Posts per user: $POSTS"
echo "  Clean existing data: $CLEAN"
echo "  Verbose output: $VERBOSE"
echo "  Environment file: $ENV_FILE"
echo ""

print_info "Estimated data to be generated:"
echo "  📊 Users: $USERS"
echo "  📝 Posts: ~$((USERS * POSTS))"
echo "  🤝 Follows: ~$((USERS * 8))"
echo "  💝 Likes: ~$((USERS * POSTS * 12))"
echo "  💬 Comments: ~$((USERS * POSTS * 4))"
echo ""

# Confirm before proceeding if cleaning
if [[ "$CLEAN" == "true" ]]; then
    print_warning "This will DELETE ALL existing data in your database!"
    read -p "Are you sure you want to continue? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_info "Operation cancelled"
        exit 0
    fi
fi

print_info "Starting data generation..."

# Set environment file
export $(cat "$ENV_FILE" | grep -v '^#' | xargs)

# Run the data generation
if go run cmd/seed/main.go $GO_ARGS; then
    print_success "Data generation completed successfully!"
    
    echo ""
    print_info "🎉 Your Social Media API is now populated with data!"
    echo ""
    print_info "Sample credentials created:"
    echo "  👑 Admin: admin@example.com / admin123"
    echo "  👤 Users: user1@example.com / password123"
    echo "           user2@example.com / password123"
    echo "           ... (user{N}@example.com / password123)"
    echo ""
    print_info "You can now:"
    echo "  🚀 Start your API server: go run cmd/server/main.go"
    echo "  📖 Check API docs at: http://localhost:8080/api/v1/"
    echo "  🔍 View health at: http://localhost:8080/health"
    echo "  🛠️ Access dev tools at: http://localhost:8080/dev/"
    
else
    print_error "Data generation failed!"
    print_info "Please check the error messages above and:"
    echo "  1. Ensure your database is running"
    echo "  2. Check your .env configuration"
    echo "  3. Verify database connection settings"
    exit 1
fi

# Optional: Show database stats
if command -v mongosh &> /dev/null || command -v mongo &> /dev/null; then
    echo ""
    read -p "Would you like to see database statistics? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        print_info "Database statistics:"
        
        # Create a small script to show stats
        cat << 'EOF' > /tmp/db_stats.js
const collections = ['users', 'posts', 'comments', 'likes', 'follows', 'stories', 'groups'];
collections.forEach(coll => {
    const count = db[coll].countDocuments();
    print(`  📊 ${coll}: ${count} documents`);
});
EOF
        
        # Try mongosh first, then mongo
        if command -v mongosh &> /dev/null; then
            mongosh "$DATABASE_URI/$DATABASE_NAME" --quiet /tmp/db_stats.js
        elif command -v mongo &> /dev/null; then
            mongo "$DATABASE_URI/$DATABASE_NAME" --quiet /tmp/db_stats.js
        fi
        
        rm -f /tmp/db_stats.js
    fi
fi

print_success "Setup complete! Happy coding! 🚀"