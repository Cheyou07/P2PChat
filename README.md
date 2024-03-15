## P2PChat (Computer Networks [UE22CS252B] Mini Project)

To set up and run the P2PChat application, follow these instructions:

### Prerequisites

- Go language installed on your system (can be installed from the [official Go website](https://golang.org/doc/install)).
- Git installed on your system (optional, for cloning the repository).

### Installation

1. Clone the repository to your local machine:

   ```bash
   git clone https://github.com/your_username/P2PChat.git
   ```

   Alternatively, you can download the ZIP file and extract it to your desired location.

2. Navigate to the project directory:

   ```bash
   cd P2PChat
   ```

3. Ensure that required modules are installed:

   ```bash
   go mod tidy
   ```

### Build and Run

#### Using Makefile

1. Run the following command to build the application using the Makefile:

   ```bash
   make
   ```

2. Once the build process is complete, execute the compiled binary:

   ```bash
   ./p2pchat
   ```

#### Without Makefile

Alternatively, you can directly run the application without using the Makefile:

1. Build and run the application using the following command:

   ```bash
   go run cmd/main.go
   ```

### Usage

Once the application is running, you can interact with the P2PChat interface to send and receive messages in real-time. You will also have the ability to name and create your own rooms. Chatting is restricted to your local network.

### Acknowledgements

This project builds upon the libp2p Go library, leveraging its powerful features for peer-to-peer communication. The go-libp2p library along with its examples served as the foundation for this project.
