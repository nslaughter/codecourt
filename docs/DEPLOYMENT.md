# CodeCourt Deployment Guide

This document provides detailed instructions for deploying CodeCourt to various environments, with a focus on production deployments.

## Deployment Options

CodeCourt can be deployed to any Kubernetes cluster using Helm charts. This guide covers:

1. Production deployment considerations
2. Resource requirements
3. Security configuration
4. Monitoring and observability
5. Backup and disaster recovery

## Prerequisites

Before deploying CodeCourt to production, ensure you have:

- Kubernetes cluster (v1.21+)
- Helm 3.x installed
- kubectl configured to access your cluster
- Persistent storage available for PostgreSQL
- Domain name and TLS certificates for secure access

## Deployment Architecture

A production deployment of CodeCourt consists of:

- API Gateway service with multiple replicas
- Microservices (User, Problem, Submission, Judging, Notification)
- PostgreSQL database (either managed or deployed via operator)
- Kafka cluster (either managed or deployed via operator)
- Ingress controller for external access
- Monitoring and logging stack

## Kubernetes Resource Requirements

Minimum recommended resources for a production deployment:

| Component | Replicas | CPU Request | Memory Request | Storage |
|-----------|----------|-------------|----------------|---------|
| API Gateway | 2 | 500m | 512Mi | - |
| User Service | 2 | 500m | 512Mi | - |
| Problem Service | 2 | 500m | 512Mi | - |
| Submission Service | 2 | 500m | 512Mi | - |
| Judging Service | 2 | 1000m | 1Gi | - |
| Notification Service | 2 | 500m | 512Mi | - |
| PostgreSQL | 3 | 1000m | 2Gi | 20Gi |
| Kafka | 3 | 1000m | 2Gi | 20Gi |
| Zookeeper | 3 | 500m | 1Gi | 10Gi |

## Deployment Steps

### 1. Prepare Namespace and RBAC

```bash
# Create namespace
kubectl create namespace codecourt

# Create service accounts and RBAC roles (if not using Helm)
kubectl apply -f k8s/rbac.yaml
```

### 2. Configure Values

Create a custom values file (`production-values.yaml`) with your production settings:

```yaml
global:
  environment: production
  storageClass: "your-storage-class"
  domain: "codecourt.example.com"

postgresql:
  enabled: true
  persistence:
    size: 20Gi
  resources:
    requests:
      cpu: 1000m
      memory: 2Gi
  
kafka:
  enabled: true
  persistence:
    size: 20Gi
  resources:
    requests:
      cpu: 1000m
      memory: 2Gi

apiGateway:
  replicaCount: 2
  resources:
    requests:
      cpu: 500m
      memory: 512Mi
  ingress:
    enabled: true
    annotations:
      kubernetes.io/ingress.class: nginx
      cert-manager.io/cluster-issuer: letsencrypt-prod
    hosts:
      - host: codecourt.example.com
        paths:
          - path: /
            pathType: Prefix
    tls:
      - secretName: codecourt-tls
        hosts:
          - codecourt.example.com

# Configure other services similarly
```

### 3. Install with Helm

```bash
# Add the CodeCourt Helm repository
helm repo add codecourt https://nslaughter.github.io/codecourt/charts
helm repo update

# Install CodeCourt
helm install codecourt codecourt/codecourt \
  --namespace codecourt \
  --values production-values.yaml \
  --timeout 10m
```

### 4. Verify Deployment

```bash
# Check pod status
kubectl get pods -n codecourt

# Check services
kubectl get svc -n codecourt

# Check ingress
kubectl get ingress -n codecourt
```

## Security Configuration

### TLS Configuration

Secure your deployment with TLS:

```yaml
apiGateway:
  ingress:
    enabled: true
    annotations:
      cert-manager.io/cluster-issuer: letsencrypt-prod
    tls:
      - secretName: codecourt-tls
        hosts:
          - codecourt.example.com
```

### Secret Management

For production, use a secrets management solution like Kubernetes Secrets, HashiCorp Vault, or cloud provider secret stores.

Update the values file to reference existing secrets:

```yaml
apiGateway:
  existingSecret: "codecourt-api-gateway-secrets"

userService:
  existingSecret: "codecourt-user-service-secrets"

# Configure other services similarly
```

### Network Policies

Enable network policies to restrict traffic between services:

```yaml
global:
  networkPolicies:
    enabled: true
```

## High Availability Configuration

For high availability:

1. Deploy multiple replicas of each service
2. Use Pod Disruption Budgets
3. Distribute across multiple nodes/zones
4. Configure proper health checks and readiness probes

Example configuration:

```yaml
apiGateway:
  replicaCount: 3
  podDisruptionBudget:
    enabled: true
    minAvailable: 2
  affinity:
    podAntiAffinity:
      preferredDuringSchedulingIgnoredDuringExecution:
        - weight: 100
          podAffinityTerm:
            labelSelector:
              matchExpressions:
                - key: app.kubernetes.io/component
                  operator: In
                  values:
                    - api-gateway
            topologyKey: kubernetes.io/hostname
```

## Monitoring and Observability

### Prometheus and Grafana

Enable Prometheus metrics:

```yaml
global:
  metrics:
    enabled: true
    serviceMonitor:
      enabled: true
```

### Logging

Configure logging to your preferred solution (ELK, Loki, etc.):

```yaml
global:
  logging:
    enabled: true
    format: json
```

## Scaling Considerations

### Horizontal Pod Autoscaling

Enable autoscaling for services:

```yaml
apiGateway:
  autoscaling:
    enabled: true
    minReplicas: 2
    maxReplicas: 5
    targetCPUUtilizationPercentage: 80
    targetMemoryUtilizationPercentage: 80
```

### Vertical Pod Autoscaling

Consider using Vertical Pod Autoscaler for optimizing resource requests.

## Backup and Disaster Recovery

### Database Backups

Configure regular PostgreSQL backups:

```yaml
postgresql:
  backup:
    enabled: true
    schedule: "0 2 * * *"  # Daily at 2 AM
    destination: "s3://your-bucket/codecourt/backups"
```

### Disaster Recovery Plan

1. **Regular Backups**: Ensure database backups are taken regularly
2. **Backup Verification**: Periodically verify backup integrity
3. **Recovery Testing**: Practice recovery procedures
4. **Documentation**: Maintain detailed recovery procedures

## Upgrade Procedures

To upgrade CodeCourt:

```bash
# Update Helm repositories
helm repo update

# Upgrade CodeCourt
helm upgrade codecourt codecourt/codecourt \
  --namespace codecourt \
  --values production-values.yaml \
  --timeout 10m
```

### Rollback Procedures

If an upgrade fails:

```bash
# List Helm releases
helm history codecourt -n codecourt

# Rollback to previous version
helm rollback codecourt 1 -n codecourt
```

## Production Checklist

Before going live, ensure:

- [ ] All secrets are properly managed
- [ ] TLS is configured for all ingress points
- [ ] Resource requests and limits are set appropriately
- [ ] Monitoring and alerting are configured
- [ ] Backup procedures are tested
- [ ] High availability is configured
- [ ] Network policies are in place
- [ ] Logging is configured and accessible
- [ ] Upgrade and rollback procedures are documented and tested

## Troubleshooting

### Common Issues

1. **Pod Startup Failures**:
   - Check pod logs: `kubectl logs -n codecourt <pod-name>`
   - Check events: `kubectl get events -n codecourt`

2. **Database Connection Issues**:
   - Verify secrets are correctly mounted
   - Check network policies
   - Ensure PostgreSQL is running

3. **Kafka Connection Issues**:
   - Verify Kafka cluster status
   - Check network policies
   - Ensure topics are created

## Conclusion

This deployment guide covers the essential aspects of deploying CodeCourt to production. For specific cloud provider instructions or advanced configurations, please refer to the provider-specific documentation or contact the CodeCourt team for assistance.
