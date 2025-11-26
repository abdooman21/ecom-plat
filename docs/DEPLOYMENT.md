# Production Deployment Guide

## 🚀 Deployment Options

### Option 1: Docker Compose (Simple)

Best for: Small to medium deployments, single-server setups

```bash
# 1. Clone repository on server
git clone https://github.com/abdooman21/ecom-plat.git
cd ecom-plat

# 2. Create production environment file
cat > .env << EOF
RABBITMQ_URL=amqp://admin:STRONG_PASSWORD@rabbitmq:5672/
APP_ENV=production
SERVICE_NAME=ecom-platform
LOG_LEVEL=info
RABBITMQ_RECONNECT_DELAY=5s
RABBITMQ_MAX_RECONNECT=10
RABBITMQ_PREFETCH_COUNT=20
SHUTDOWN_TIMEOUT=30s
EOF

# 3. Start services
docker-compose up -d

# 4. Scale consumers based on load
docker-compose up -d --scale consumer=5

# 5. Monitor logs
docker-compose logs -f
```

### Option 2: Kubernetes (Recommended for Production)

Create Kubernetes manifests:

```yaml
# k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: ecom-platform

---
# k8s/rabbitmq-statefulset.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: rabbitmq
  namespace: ecom-platform
spec:
  serviceName: rabbitmq
  replicas: 3
  selector:
    matchLabels:
      app: rabbitmq
  template:
    metadata:
      labels:
        app: rabbitmq
    spec:
      containers:
      - name: rabbitmq
        image: rabbitmq:3.12-management-alpine
        ports:
        - containerPort: 5672
        - containerPort: 15672
        env:
        - name: RABBITMQ_DEFAULT_USER
          valueFrom:
            secretKeyRef:
              name: rabbitmq-secret
              key: username
        - name: RABBITMQ_DEFAULT_PASS
          valueFrom:
            secretKeyRef:
              name: rabbitmq-secret
              key: password
        volumeMounts:
        - name: rabbitmq-data
          mountPath: /var/lib/rabbitmq
  volumeClaimTemplates:
  - metadata:
      name: rabbitmq-data
    spec:
      accessModes: [ "ReadWriteOnce" ]
      resources:
        requests:
          storage: 10Gi

---
# k8s/producer-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: producer
  namespace: ecom-platform
spec:
  replicas: 2
  selector:
    matchLabels:
      app: producer
  template:
    metadata:
      labels:
        app: producer
    spec:
      containers:
      - name: producer
        image: your-registry/ecom-producer:latest
        env:
        - name: RABBITMQ_URL
          valueFrom:
            secretKeyRef:
              name: rabbitmq-secret
              key: url
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "200m"

---
# k8s/consumer-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: consumer
  namespace: ecom-platform
spec:
  replicas: 5
  selector:
    matchLabels:
      app: consumer
  template:
    metadata:
      labels:
        app: consumer
    spec:
      containers:
      - name: consumer
        image: your-registry/ecom-consumer:latest
        env:
        - name: RABBITMQ_URL
          valueFrom:
            secretKeyRef:
              name: rabbitmq-secret
              key: url
        - name: RABBITMQ_PREFETCH_COUNT
          value: "20"
        resources:
          requests:
            memory: "256Mi"
            cpu: "200m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          exec:
            command:
            - /bin/sh
            - -c
            - pgrep consumer
          initialDelaySeconds: 30
          periodSeconds: 10

---
# k8s/hpa.yaml (Horizontal Pod Autoscaler)
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: consumer-hpa
  namespace: ecom-platform
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: consumer
  minReplicas: 3
  maxReplicas: 20
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

Deploy to Kubernetes:
```bash
# Create secret
kubectl create secret generic rabbitmq-secret \
  --from-literal=username=admin \
  --from-literal=password=STRONG_PASSWORD \
  --from-literal=url=amqp://admin:STRONG_PASSWORD@rabbitmq:5672/ \
  -n ecom-platform

# Apply manifests
kubectl apply -f k8s/
```

## 🔒 Security Checklist

### RabbitMQ Security

1. **Change default credentials**
```bash
# In docker-compose.yml or Kubernetes secrets
RABBITMQ_DEFAULT_USER=admin
RABBITMQ_DEFAULT_PASS=<generate-strong-password>
```

2. **Enable TLS/SSL**
```yaml
# rabbitmq.conf
listeners.ssl.default = 5671
ssl_options.cacertfile = /path/to/ca_certificate.pem
ssl_options.certfile   = /path/to/server_certificate.pem
ssl_options.keyfile    = /path/to/server_key.pem
ssl_options.verify     = verify_peer
ssl_options.fail_if_no_peer_cert = true
```

3. **Configure virtual hosts**
```bash
rabbitmqctl add_vhost production
rabbitmqctl set_permissions -p production admin ".*" ".*" ".*"
```

4. **Enable authentication plugins**
```bash
rabbitmq-plugins enable rabbitmq_auth_backend_ldap
```

### Application Security

1. **Use environment variables for secrets** (Never hardcode)
2. **Enable rate limiting** on producers
3. **Validate message payloads** in consumers
4. **Implement message signing** for critical data
5. **Use network policies** in Kubernetes

## 📊 Monitoring Setup

### Prometheus + Grafana

```yaml
# docker-compose.monitoring.yml
version: '3.8'

services:
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana_data:/var/lib/grafana

  rabbitmq-exporter:
    image: kbudde/rabbitmq-exporter
    ports:
      - "9419:9419"
    environment:
      - RABBIT_URL=http://rabbitmq:15672
      - RABBIT_USER=admin
      - RABBIT_PASSWORD=admin123

volumes:
  prometheus_data:
  grafana_data:
```

### Key Metrics to Monitor

1. **RabbitMQ Metrics**
   - Queue length
   - Message rate (publish/consume)
   - Connection count
   - Memory usage
   - Disk space

2. **Application Metrics**
   - Orders processed/failed
   - Processing time (p50, p95, p99)
   - Error rate
   - Retry count
   - Memory/CPU usage

3. **System Metrics**
   - Container CPU/Memory
   - Network I/O
   - Disk I/O

### Alerts Configuration

```yaml
# prometheus/alerts.yml
groups:
- name: rabbitmq
  rules:
  - alert: HighQueueLength
    expr: rabbitmq_queue_messages > 10000
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "High queue length detected"
      
  - alert: ConsumerDown
    expr: up{job="consumer"} == 0
    for: 2m
    labels:
      severity: critical
    annotations:
      summary: "Consumer instance is down"
      
  - alert: HighErrorRate
    expr: rate(orders_failed[5m]) > 0.1
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "High error rate detected"
```

## 🔄 CI/CD Pipeline

### GitHub Actions (Already configured in .github/workflows/ci.yml)

Additional deployment workflow:

```yaml
# .github/workflows/deploy.yml
name: Deploy to Production

on:
  push:
    tags:
      - 'v*'

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    
    - name: Build and push images
      # ... docker build steps
    
    - name: Deploy to Kubernetes
      uses: azure/k8s-deploy@v4
      with:
        manifests: |
          k8s/producer-deployment.yaml
          k8s/consumer-deployment.yaml
        images: |
          your-registry/ecom-producer:${{ github.sha }}
          your-registry/ecom-consumer:${{ github.sha }}
```

## 🔧 Performance Tuning

### RabbitMQ Configuration

```conf
# rabbitmq.conf
vm_memory_high_watermark.relative = 0.6
disk_free_limit.absolute = 5GB
heartbeat = 60
channel_max = 2047
```

### Consumer Tuning

```go
// Adjust based on message processing time
RABBITMQ_PREFETCH_COUNT=20  // Lower for heavy processing
RABBITMQ_PREFETCH_COUNT=100 // Higher for light processing
```

### Scaling Guidelines

| Load Level | Consumers | Prefetch | RabbitMQ Memory |
|------------|-----------|----------|-----------------|
| Low | 2-3 | 10 | 2GB |
| Medium | 5-10 | 20 | 4GB |
| High | 10-20 | 30 | 8GB |
| Very High | 20+ | 50 | 16GB+ |

## 📝 Backup Strategy

### RabbitMQ Definitions Backup

```bash
# Export definitions
rabbitmqadmin export definitions.json

# Schedule daily backups
0 2 * * * /usr/local/bin/rabbitmqadmin export /backups/definitions-$(date +\%Y\%m\%d).json
```

### Message Persistence

Ensure messages are persisted:
```go
amqp.Publishing{
    DeliveryMode: amqp.Persistent, // This is critical!
    ContentType:  "application/json",
    Body:         body,
}
```

## 🚨 Incident Response

### Common Issues

1. **Queue buildup**
```bash
# Check queue status
rabbitmqctl list_queues name messages consumers

# Scale consumers
kubectl scale deployment consumer --replicas=10
```

2. **Memory pressure**
```bash
# Check RabbitMQ memory
rabbitmqctl status | grep memory

# Increase vm_memory_high_watermark if needed
```

3. **Connection issues**
```bash
# Check connections
rabbitmqctl list_connections

# Close problematic connections
rabbitmqctl close_connection "<connection.name>" "reason"
```

## 📞 Support & Maintenance

### Health Checks

```bash
# RabbitMQ health
rabbitmqctl node_health_check

# Application health
curl http://localhost:8080/health  # If health endpoint added
```

### Log Aggregation

Use ELK stack or similar:
```yaml
# docker-compose.logging.yml
services:
  elasticsearch:
    image: elasticsearch:8.8.0
    
  logstash:
    image: logstash:8.8.0
    
  kibana:
    image: kibana:8.8.0
    ports:
      - "5601:5601"
```

## 📚 Additional Resources

- [RabbitMQ Best Practices](https://www.rabbitmq.com/production-checklist.html)
- [Go Best Practices](https://go.dev/doc/effective_go)
- [Kubernetes Production Best Practices](https://kubernetes.io/docs/setup/best-practices/)