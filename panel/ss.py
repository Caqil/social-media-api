import os

def create_file(path):
    """Create an empty file at the specified path."""
    with open(path, 'w') as f:
        pass

def create_app_directory_structure():
    """Create the app directory structure for the admin panel project."""
    # Root app directory
    app_dir = os.path.join("panel", "app")
    os.makedirs(app_dir, exist_ok=True)
    
    # app root files
    app_files = [
        "layout.tsx",
        "page.tsx",
        "loading.tsx",
        "error.tsx",
        "not-found.tsx",
        "globals.css"
    ]
    for file in app_files:
        create_file(os.path.join(app_dir, file))
    
    # app/api directory
    api_dir = os.path.join(app_dir, "api")
    os.makedirs(os.path.join(api_dir, "auth"), exist_ok=True)
    create_file(os.path.join(api_dir, "auth", "route.ts"))
    os.makedirs(os.path.join(api_dir, "upload"), exist_ok=True)
    create_file(os.path.join(api_dir, "upload", "route.ts"))
    
    # app/(auth) directory
    auth_dir = os.path.join(app_dir, "(auth)")
    auth_subdirs = [
        "login",
        "forgot-password",
        "reset-password"
    ]
    for subdir in auth_subdirs:
        os.makedirs(os.path.join(auth_dir, subdir), exist_ok=True)
        create_file(os.path.join(auth_dir, subdir, "page.tsx"))
    create_file(os.path.join(auth_dir, "layout.tsx"))
    
    # app/(dashboard) directory
    dashboard_dir = os.path.join(app_dir, "(dashboard)")
    os.makedirs(dashboard_dir, exist_ok=True)
    create_file(os.path.join(dashboard_dir, "layout.tsx"))
    
    # app/(dashboard) subdirectories
    dashboard_subdirs = [
        ("dashboard", ["page.tsx"]),
        ("users", ["page.tsx", "loading.tsx"]),
        ("posts", ["page.tsx", "loading.tsx"]),
        ("comments", ["page.tsx", "loading.tsx"]),
        ("stories", ["page.tsx", "loading.tsx"]),
        ("groups", ["page.tsx", "loading.tsx"]),
        ("messages", ["page.tsx", "loading.tsx"]),
        ("reports", ["page.tsx", "loading.tsx"]),
        ("media", ["page.tsx", "loading.tsx"]),
        ("analytics", ["page.tsx", "loading.tsx"]),
        ("moderation", ["page.tsx", "loading.tsx"]),
        ("notifications", ["page.tsx", "loading.tsx"]),
        ("settings", ["page.tsx", "loading.tsx"])
    ]
    
    for subdir, files in dashboard_subdirs:
        subdir_path = os.path.join(dashboard_dir, subdir)
        os.makedirs(subdir_path, exist_ok=True)
        for file in files:
            create_file(os.path.join(subdir_path, file))

if __name__ == "__main__":
    create_app_directory_structure()
    print("App directory structure created successfully!")