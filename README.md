# sqs-alerter

This is a tool to alert about messages available in a SQS queue - usually a dead letter queue. Alerts are currently sent to Slack only

## Running

Just build the tool using `make` and then run it

```bash
make
```

```bash
./sqs-alerter
```

`sqs-alerter` is just a simple tool and does not run as a service, it just runs once and then exits. To keep it running, run it as a cron job or using `watch` command continuously every few seconds or minutes, whatever interval you wish

### Setup

#### AWS credentials

Create an IAM user which has access to `sqs:GetQueueAttributes` and `sqs:ReceiveMessage`
