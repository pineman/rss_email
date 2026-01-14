# Fly.io Deployment Notes

## Current Status
- **App Name**: rss-email
- **Region**: cdg (Paris)
- **Machine Count**: 1 (as intended)
- **Machine Type**: shared-cpu-1x with 256MB RAM
- **Volume**: vol_4o5wn97k0dwgyk2v (rss_email_data) mounted at `/app/data`
- **Volume Count**: 1 (as intended)

## Best Practices for Single-Machine Apps

### Deployment Strategy
The `fly.toml` now uses `strategy = "immediate"` which:
- Replaces the existing machine directly during deployments
- Avoids creating temporary additional machines
- Prevents the "2 machines" issue from recurring

### Manual Deployments
When deploying manually, use:
```bash
flyctl deploy
```

This will respect the `strategy = "immediate"` setting and replace the existing machine.

### Checking Machine and Volume Count
To verify you have only one machine and one volume:
```bash
flyctl machine list
flyctl volumes list
```

Expected output:
- **1 machine** in "started" state
- **1 volume** in "created" state, attached to your machine

### Scaling Considerations
If you ever need to scale to multiple machines:
```bash
# Scale up (creates additional machines)
flyctl scale count 2

# Scale down (removes machines)
flyctl scale count 1
```

However, for this RSS email worker, **1 machine should be sufficient** since:
- It's a background worker processing RSS feeds periodically
- The volume is attached to a single machine
- No need for redundancy or load balancing

## Volume Management
The app uses a persistent volume:
- **ID**: `vol_4o5wn97k0dwgyk2v`
- **Name**: `rss_email_data`
- **Mount**: `/app/data`
- **Size**: 1GB
- **Purpose**: Stores RSS feed state and email history

**Important Notes**:
- Volumes are tied to specific machines and regions
- When a machine is destroyed, its volume becomes orphaned (not automatically deleted)
- Always check for orphaned volumes after destroying machines: `flyctl volumes list`
- Delete orphaned volumes to avoid unnecessary costs: `flyctl volumes destroy <volume_id>`

### Preventing Duplicate Volumes
With `deploy.strategy = "immediate"`, Fly.io will:
- Replace the existing machine in-place
- Keep the same volume attached
- NOT create new volumes during deployments

This prevents the accumulation of orphaned volumes.

## Monitoring
Check app status:
```bash
flyctl status          # Overall app status
flyctl machine list    # List all machines
flyctl logs            # View application logs
```
