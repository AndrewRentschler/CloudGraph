# CloudGraph

CloudGraph is a distributed network monitoring tool designed to create a real-time graph of network latency across multiple cloud providers. By deploying lightweight Go-based agents (nodes) on the smallest instances available from various cloud providers, CloudGraph collects and centralizes ping times between these nodes, enabling you to visualize the connectivity and performance across the entire network.

## Overview

The CloudGraph system consists of two main components:

1. **Central Server**: Hosts the frontend and handles the ORM (Object-Relational Mapping). It also exposes a GraphQL API for querying the network graph and visualizing the web of nodes and their respective ping times.

2. **Node Agents**: These are lightweight Go scripts deployed on small instances across different cloud providers. Each node registers itself with the central server, receives a unique `nodeId` and an array of IP addresses to ping. The node then continuously pings these IPs and sends the results back to the central server.

## How It Works

1. **Node Registration**: On boot, each node agent sends a request to the central server to register itself. The server responds with a unique `nodeId` and a list of IP addresses (representing other nodes) that this node should ping.

2. **Ping Loop**: The node enters a loop where it pings each IP address provided by the central server. It measures the latency for each ping and stores the results.

3. **Reporting Results**: After completing the ping loop, the node sends the results back to the central server, which collects and stores the data.

4. **Centralized GraphQL API**: The central server hosts a GraphQL API that allows users to query the network graph, visualizing the ping times between all nodes.

## Installation

### Central Server

1. Set up a server to host the frontend and the GraphQL API.
2. Deploy your preferred ORM to manage the database.
3. Ensure the server is accessible to all deployed nodes.

### Node Agent

1. Clone the repository:
    ```bash
    git clone https://github.com/AndrewRentschler/CloudGraph.git
    cd CloudGraph
    ```

2. Set the `CENTRAL_SERVER_IP` environment variable to point to your central server's IP address:
    ```bash
    export CENTRAL_SERVER_IP="your-central-server-ip"
    ```

3. Build and run the Go script:
    ```bash
    go build -o cloudgraph-node
    ./cloudgraph-node
    ```

4. Deploy this node agent on the smallest available instance from your cloud provider.

## Usage

Once the central server and node agents are set up and running:

- The central server will maintain a real-time graph of network latency between all nodes.
- You can query the GraphQL API to visualize the network graph and view ping times between any two nodes.

### Example GraphQL Query

```graphql
{
  nodes {
    id
    pings {
      targetNode
      latency
    }
  }
}
```

## Contributing

Contributions are welcome! Please fork the repository and submit a pull request with your changes.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Thanks to the open-source community for providing the tools and libraries that made this project possible.

## Contact

For any questions or suggestions, please open an issue.