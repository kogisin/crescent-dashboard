# Crescent Dashboard

## Prerequisites

### Prometheus

```bash
# Install using homebrew
brew install prometheus

# Start prometheus
brew services start prometheus

# Stop prometheus
brew services stop prometheus

# Restart prometheus
brew services restart prometheus

# Config file location
# Homebrew v2: /usr/local/etc/prometheus.yml
# Homebrew v3: /opt/homebrew/etc/prometheus.yml

# Binary location (prometheus, promtool, prometheus_brew_services)
# Homebrew v2: /usr/local/bin/prometheus
# Homebrew v2: /opt/homebrew/opt/prometheus/bin
```

### Grafana

```bash
# Install
brew install grafana

# Start
brew services start grafana

# Stop
brew services stop grafana

# Restart
brew services restart grafana

# Config file location
# Homebrew v2: /usr/local/etc/grafana/grafana.ini
# Homebrew v3: /opt/homebrew/etc/grafana/grafana.ini

# Binary location (grafana-cli, grafana-server)
# Homebrew v2: /usr/local/Cellar/grafana/[version]
# Homebrew v3: /opt/homebrew/Cellar/grafana/[version]

# Debug if localhost:3000 web interface does not display dashboard
# Monitor grafana proceess by running "$ps aux | grep grafana"
# 1. Tailing log: tail -f /usr/local/var/log/grafana/grafana.log
# 2. Try change port from default 3000 to 3001 and restart grafana
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
vim /usr/local/etc/prometheus.yml (Homebrew v2)
vim /opt/homebrew/etc/prometheus.yml (Homebrew v3)

# 1. Go to Grafana client http://localhost:3000/login (admin/admin)
# 2. Configuration -> Data Sources -> Add prometheus datasource http://localhost:9090 
# 3. Create -> Import -> Upload JSON file (/dashboard/Crescent Dashboard.....json) -> Select Prometheus (default)
# 4. You will be able to see the dashboard. Enjoy!
```
