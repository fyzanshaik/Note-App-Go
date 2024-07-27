# 🚀 Go-NoteApp: A Simple Note-Taking Web Application

Welcome to Go-NoteApp, a straightforward and efficient note-taking web application built with Go! This application allows you to create, view, edit, and manage your notes effortlessly. Whether you're
a developer, student, or professional, you'll find Go-NoteApp to be a handy tool for organizing your thoughts and ideas. 📝

## Features

-  **📝 Create and Edit Notes**: Easily create new notes and edit existing ones.
-  **👀 View Notes**: View the contents of your notes in a user-friendly interface.
-  **💼 Manage Notes**: Organize and manage your notes efficiently.
-  **🔐 Secure**: Ensure the security of your notes with local storage.
-  **🌐 Cross-Platform**: Access your notes from any device with a web browser.

## Getting Started

To get started with Go-NoteApp, follow these steps:

### 1. Clone the Repository

```bash
git clone https://github.com/your-username/go-noteapp.git
```

### 2. Navigate to Project Directory

```bash
cd go-noteapp
```

### 3. Install Development Dependencies

For development, you need to install Air for live-reloading. Run:

```bash
go install github.com/cosmtrek/air@latest
```

### 4. Create Air Configuration File (Optional)

Create a `.air.toml` file in the root of your project to customize Air’s behavior:

```toml
# .air.toml

[build]
bin = "./bin/main"  # Path to your build output binary
cmd = "go build -o ${build.bin} ."  # Command to build your application
include = ["*.go", "templates/**/*.html"]  # Files to watch for changes
exclude = ["static"]  # Files to exclude from watching

[logger]
time = true  # Enable logging with timestamps
```

### 5. Run the Application in Development Mode

Use Air to start the application with live-reloading:

```bash
air
```

This will start the application and automatically restart the server when changes are detected.

### 6. Run the Application in Production Mode

For production, build your application and run the resulting binary:

```bash
go build -o go-noteapp
./go-noteapp
```

Open your web browser and go to `http://localhost:3000` to access the application.

## Usage

### Creating a Note

1. Click on the "Create a note!" button on the homepage.
2. Enter a title and the content of your note in the provided form.
3. Click on the "Create Note" button to save your note.

### Viewing a Note

1. Click on the title of the note you want to view from the homepage.
2. You will be directed to a page where you can see the title and content of the selected note.

### Editing a Note

1. Click on the "edit" link next to the title of the note you want to edit from the homepage.
2. You will be directed to a page where you can edit the title and content of the note.
3. After making your changes, click on the "Save" button to update the note.

## Deployment

To deploy this application to a server or platform of your choice:

1. **Build the application**:

   ```bash
   go build -o go-noteapp
   ```

2. **Deploy the `go-noteapp` binary** to your server or platform.

3. **Run the binary** on your server:

   ```bash
   ./go-noteapp
   ```

   Ensure the server is set to listen on the appropriate port and is configured for your deployment environment.

## License

Distributed under the MIT License. See `LICENSE` for more information.

---

Feel free to explore the features of Go-NoteApp and customize it according to your needs. Happy note-taking! 🚀
