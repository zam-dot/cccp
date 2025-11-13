// test_safe_copy.tpl - FIXED VERSION

// Includes first
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <fcntl.h>
#include <sys/stat.h>

// Macros (choose ONE auto-cleanup system)
{{ "" | generate_error_macros }}
{{ "" | generate_auto_cleanup }}

void copy_file_safely(const char* source_path, const char* dest_path) {
    // Validate input pointers
    {{ "source_path" | check_null : "source path" }}
    {{ "dest_path" | check_null : "destination path" }}
    
    // Open source file
    int source_fd = open(source_path, O_RDONLY);
    {{ "source_fd" | check_syscall : "opening source file" }}
    
    // Get file size
    struct stat st;
    {{ "fstat(source_fd, &st)" | check_syscall : "getting file stats" }};
    size_t file_size = st.st_size;
    
    // Allocate buffer - ONLY ONCE!
    AUTO_FREE char* buffer;
    {{ "buffer" | get_memory : "file_size + 1" }}
    
    // Read file
    ssize_t bytes_read = read(source_fd, buffer, file_size);
    {{ "bytes_read" | check_syscall : "reading source file" }}
    
    if (bytes_read < 0) {
        perror("read failed");
        exit(EXIT_FAILURE);
    }
    
    buffer[bytes_read] = '\0';
    
    // Create destination file
    int dest_fd = open(dest_path, O_WRONLY | O_CREAT | O_TRUNC, 0644);
    {{ "dest_fd" | check_syscall : "creating destination file" }}
    
    // Write data  
    ssize_t bytes_written = write(dest_fd, buffer, bytes_read);
    {{ "bytes_written" | check_syscall : "writing destination file" }}
    
    if (bytes_written != bytes_read) {
        fprintf(stderr, "Write incomplete: wrote %zd of %zd bytes\n", bytes_written, bytes_read);
        exit(EXIT_FAILURE);
    }
    
    // Close files
    {{ "close(source_fd)" | check_syscall : "closing source file" }};
    {{ "close(dest_fd)" | check_syscall : "closing destination file" }};
    
    printf("Successfully copied %s to %s (%zd bytes)\n", source_path, dest_path, bytes_written);
}

int main(int argc, char* argv[]) {
    {{ "argc != 3" | check_args : "wrong number of arguments" }}     
    const char* source = argv[1];
    const char* dest = argv[2];
    
    copy_file_safely(source, dest);
    return 0;
}
