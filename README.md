# Crescent Dashboard

## Prerequisites

### Prometheus

```bash
# Install using homebrew
brew install prometheus

# Config file location
/opt/homebrew/etc/prometheus.yml

# Binary location (prometheus, promtool, prometheus_brew_services)
/opt/homebrew/opt/prometheus/bin

# Start prometheus
brew services start prometheus

# Restart prometheus
brew services restart prometheus
```

### Grafana

```bash
# Install
brew install grafana

# Config file location
/opt/homebrew/etc/grafana/grafana.ini

# Binary location (grafana-cli, grafana-server)
# Homebrew v2: /usr/local/Cellar/grafana/[version]
# Homebrew v3: /opt/homebrew/Cellar/grafana/[version]
/opt/homebrew/Cellar/grafana/8.4.6/bin/

# Start
brew services start grafana

# Restart
brew services restart grafana
```

## Usage

```bash
# Start prometheus and grafana
brew services start prometheus
brew services start grafana

# Run the crescent dashboard program
# Prometheus metrics: http://localhost:9090/metrics
# Custom metrics: http://localhost:2112/metrics
go run . mainnet.crescent.network:9090 https://apigw.crescent.network/

# Add localhost:2112 to let prometheus to scrape custom metrics
vim /opt/homebrew/etc/prometheus.yml

# 1. Go to Grafana client http://localhost:3000/login (admin/admin)
# 2. Configuration -> Data Sources -> Add prometheus datasource http://localhost:9090 
# 3. Create -> Import -> Upload JSON file (/dashboard/Crescent Dashboard.....json) -> Select Prometheus (default)
# 4. You will be able to see the dashboard. Enjoy!
```
