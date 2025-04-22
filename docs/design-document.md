# Design Document

This documents outlines my design and implementation of a trading chart service implemented in Go that connects to the Binance API to fetch live tick data for specified trading pairs and process the data by aggregating it into OHLC (Open, High, Low, Close) candlesticks over configurable intervals. As new candlesticks are formed, the service broadcasts the current bar in real time to a streaming server API and persists completed bars to a Timescale database for future analysis and historical reference.

The service is containerised and deployed on a Kubernetes cluster, with all infrastructure provisioned using Terraform. This setup ensures consistency across environments, enables autoscaling, and supports rapid iteration and monitoring in production.

![Candle Sticks Output](/docs/images/postman%20screenshot.png "Candle Sticks Output")

##  Trade offs, Decisions and  Assumptions

I had made some assumptions and tradeoff to implement this soloution. This include the choices of tools and architecture to emply. Software architecture is largely about trade off and context; the decision on which architecture to choose depends on business drivers, the environment, and other factors

#### 1. Using Minikube for Kurbanetes  

There were no explicit requirements to use a specific cloud provider such as AWS, Azure, GCP, or Alibaba Cloud. As a result, the implementation was designed to run on **Minikube**, allowing for quick and convenient local testing and evaluation. This approach ensures  without access to cloud infrastructure you can still deploy and test my implementation. This implementation can be replicated in other clouds.

### 3. No Kafka and Decoupling with message queues
Considering the scope of this exercise and the limited real-time data—restricted to only **BTCUSDT**, **ETHUSDT**, and **PEPEUSD**—introducing a message queue system like Kafka or NATS would be excessive. For this implementation, I opted for a lightweight aggregator by using channels and goroutines, the aggregator can efficiently process a high volume of real-time data without blocking, which allows for easier scaling in the event of memory-bound spikes due to a high number of symbols. For scalability and production deployment, however, I would adopt a distributed aggregation architecture using Kafka.

### 2.  Timescale for Database
I opted for Timescale DB due to its specialized optimization for high-frequency time-series data, a critical requirement for processing real-time cryptocurrency trades. TimescaleDB's hypertables automate time-based partitioning and enable parallelized writes, scaling to handle over 1 million candlesticks per second—essential for aggregating tick data from volatile assets like BTCUSDT or PEPEUSDT. Its native columnar compression reduces storage costs by 90%+, a vital advantage given the exponential growth of OHLCV (Open/High/Low/Close/Volume) data in crypto markets. For scalability, TimescaleDB supports read replicas and distributed hypertables, ensuring low-latency query performance even during market surges. Automatic partitioning by both time and symbol aligns perfectly with the service’s dual requirements: rapid real-time aggregation (time-based chunks) and efficient per-asset historical analysis (symbol-based indexing).

##  Architecture Diagram

###  Architecture Without a message brooker (Direct Processing)
![Architecture Diagram](/docs/images/architecture.svg "Architecture Diagram")

###  For future improvements using message brooker e.g Kafka 
![Architecture Diagram](/docs/images/architecture2.svg "Architecture Diagram")

##  Data Ingestion, Aggregation, and Streaming
![Diagram](/docs/images/data-ingestions.svg "Data Ingestion, Aggregation, and Streaming")

###  Data Ingestion

I implemented a binance client service  `internal/binance/client.go` that  connects to Binance WebSocket API and consumes real-time market data (ticks) streams for the specified trading symbols. It maintains persistent connections for all symbols and implements exponential backoff for reconnections. The incoming payloads are validated and pushed into an internal channel for processing,  each symbol runs on it own goroutine and  uses a separate channel per symbol for aggregators to avoid contention. 

`tickChan := make(chan binance.Tick, cfg.Buffers.TickChan)`

The client extracts relevant fields such as Symbol , Price , Quantity , and Timestamp , converting them into a Tick struct. This struct is then sent to a channel ( tickChan ) for further processing.

The tick bufferred channel helps handle back pressure by allowing a limited number of messages to be queued in the channel before blocking the sender. The capacity is configurable and can be adjusted based on demand. The configuration is defined in `configs/config.yaml` file.

`c.connectSymbol(ctx, symbol, tickChan)`

The `connectSymbol` method spawns a separate goroutine for each trading symbol specified in the symbols slice. This allows the client to handle multiple WebSocket connections concurrently, ensuring that data for all symbols is received in real-time. Having a shared connection pool is potentially a better approach for more real time pairs but I have limited my implementation to this defined scope.

The client handles errors for connection. If a connection error occurs, the client logs the error and attempts to reconnect after a short delay.ensuring continuity in data streaming.

For sychronisation, I used `sync.WaitGroup` to wait for all goroutines to finish before closing the `tickChan` channel. This ensures that all data is processed before the channel is closed, preventing potential data loss.

###  Data Aggregation

The Aggregator implementation is designed to process real-time ticks and aggregate them into candlestick data. Aggregation is performed using a time-based sliding window. Events are grouped into 1-minute windows, with metrics computed and flushed every interval. I employed the use of  `sync.RWMutex` to manage concurrent access to shared resources, specifically the `candles` map and `lastTickTime` . This ensures thread-safe operations when multiple goroutines are reading from or writing to these resources.

The Run method continuously listens for incoming ticks from a channel (tickChan) and processes each tick to update or create candlestick data.

 `agg.Run(ctx, tickChan, candleChan)`

The processTick method aggregates ticks into candlesticks and updates the open, high, low, close, and volume fields of a Candle struct based on incoming tick data. This is essential for generating candlestick charts used in technical analysis.

Also I check for candlesticks that have reached their end time and marks them as finalised. Finalised candlesticks are sent to an output channel (candleChan) and removed from the active map. This ensures that only active candlesticks are updated with new tick data. I also handle graceful shortdown to ensure that all remaining candlesticks are finalized and sent out before the service stops.

### Data Streaming

The streaming service is built on gRPC technology, implementing a server-side streaming pattern defined in the candlestick.proto file. This architecture enables efficient one-to-many communication where a single client request initiates a continuous flow of candlestick data from the server. 

At the core of the implementation is the StreamCandlesticks method which processes client symbol requests and establishes persistent data channels. The service employs a continuous monitoring approach, checking for new candlestick data at one-second intervals for each requested symbol. When new data becomes available, it's transformed from internal Candle objects into standardized protobuf Candlestick messages with normalized time representations (Unix milliseconds) to ensure cross-platform compatibility. The system implements context-aware monitoring to gracefully handle client disconnections and prevent resource leaks.

##  Data Design

The section will explain the data model and other data structure that were employed for this project

### Tick Data
This represents individual trade events from the market with a symbol (trading pair), price, quantity, and timestamp. It maps the raw market data in the form of ticks from Binance:

```
type Tick struct {
    Symbol    string
    Price     float64
    Quantity  float64
    Timestamp time.Time
}
```

### 2. Candle Data
Ticks are aggregated into candlestick data. Candles represent price movement over a specific time period (1 minute in this implementation) with opening, high, low, and closing prices, along with volume information.

```
type Candle struct {
    Symbol    string    
    Open      float64   
    High      float64   
    Low       float64   
    Close     float64   
    Volume    float64   
    StartTime time.Time 
    EndTime   time.Time 
    Finalized bool     
}
```

### 3. Protocol Buffer Messages
For API communication, the application uses Protocol Buffers:

```
message Candlestick {
  string symbol = 1;
  double open = 2;
  double high = 3;
  double low = 4;
  double close = 5;
  double volume = 6;
  int64 start_time = 7;
  int64 end_time = 8;
  bool is_final = 9;
}
```

## Database Design
The application uses TimescaleDB, which is specialized for time-series data. Based on the struct tags in the Candle type, the database schema likely includes:

```sql
CREATE TABLE candles (
    symbol VARCHAR NOT NULL,
    open DOUBLE PRECISION NOT NULL,
    high DOUBLE PRECISION NOT NULL,
    low DOUBLE PRECISION NOT NULL,
    close DOUBLE PRECISION NOT NULL,
    volume DOUBLE PRECISION NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL
);
 ```

TimescaleDB would then convert this into a Hypertable for efficient time-series operations:

```sql
SELECT create_hypertable('candles', 'start_time');
 ```

##  Choice of K8s resources and deployment strategy

This implementation uses several standard Kubernetes resources:

 - **Deployments** :
      - trading deployment for your Go application
      - postgres deployment for the TimescaleDB database

 - **Services** :
      -  Service for the trading application exposing HTTP (8080) and gRPC (50057) ports
      - Service for the postgres database

 - **ConfigMap** :
      -  app-config containing application configuration in YAML format
      -  Mounted into the trading application container

 - **PersistentVolumeClaim** :
      - 5GB storage allocation for the TimescaleDB database
      - Ensures data persistence across pod restarts
      
 - **Ingress** 
      - HTTP ingress for web traffic on trading.local
      - gRPC-specific ingress on grpc.trading.local with appropriate annotations for gRPC protoco


### Deployment Strategy
For this implementation I focussed on the following:

1. **Simplicity and Reliability**
   - Single replica for both application and database suitable for my use case. This can be scaled to many replicas
   - Configuration externalized via ConfigMap
   - Persistent storage for database
   
2. **Service Exposure**
   - Dual protocol support (HTTP and gRPC)
   - Ingress configuration for external access
   - Proper protocol handling for gRPC traffic

3. **Configuration Management**
   - External configuration via ConfigMap
   - Volume mounting for configuration files

4. **Resource Isolation**
   - separate deployments for application and database
   - Established clear service boundaries

Given time constraints and the requirements, this is a foundation architecture with proper separation of concerns between the application and database components.

While this is fine for a start, to scale in production I will consider the following improvements: multiple replicas for high availability, resource limits and requests, horizontal pod autoscaling, network policies for enhanced security and secrets for sensitive configuration