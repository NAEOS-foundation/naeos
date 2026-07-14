package cloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/version"
)


// AWSAdapter implements CloudAdapter for Amazon Web Services.
type AWSAdapter struct {
	Runner CommandRunner
}

func (a *AWSAdapter) Name() string {
	return "AWS"
}

func (a *AWSAdapter) Provider() CloudProvider {
	return AWS
}

func (a *AWSAdapter) Validate(config *DeployConfig) error {
	if config.Region == "" {
		return fmt.Errorf("AWS region is required")
	}
	validRegions := []string{
		"us-east-1", "us-east-2", "us-west-1", "us-west-2",
		"eu-west-1", "eu-west-2", "eu-west-3",
		"eu-central-1", "eu-north-1",
		"ap-southeast-1", "ap-southeast-2", "ap-northeast-1", "ap-northeast-2",
		"ap-south-1", "sa-east-1", "ca-central-1",
	}
	valid := false
	for _, r := range validRegions {
		if config.Region == r {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid AWS region: %s", config.Region)
	}
	return nil
}

func (a *AWSAdapter) Plan(config *DeployConfig) (*PlanResult, error) {
	resources := []Resource{}
	for _, res := range config.Resources {
		switch res.Type {
		case ResourceStorage:
			resources = append(resources, Resource{
				Name: res.Name,
				Type: "aws_s3_bucket",
				Spec: map[string]any{
					"bucket": fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name),
					"region": config.Region,
				},
			})
		case ResourceCompute:
			resources = append(resources, Resource{
				Name: res.Name,
				Type: "aws_ecs_service",
				Spec: map[string]any{
					"cluster": fmt.Sprintf("%s-%s", config.Project, config.Environment),
					"service": res.Name,
				},
			})
		case ResourceDatabase:
			resources = append(resources, Resource{
				Name: res.Name,
				Type: "aws_rds_instance",
				Spec: map[string]any{
					"identifier": fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name),
					"engine":     "postgres",
				},
			})
		case ResourceCache:
			resources = append(resources, Resource{
				Name: res.Name,
				Type: "aws_elasticache_cluster",
				Spec: map[string]any{
					"cluster_id": fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name),
					"engine":     "redis",
				},
			})
		case ResourceQueue:
			resources = append(resources, Resource{
				Name: res.Name,
				Type: "aws_sqs_queue",
				Spec: map[string]any{
					"name": fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name),
				},
			})
	case ResourceCDN:
		resources = append(resources, Resource{
			Name: res.Name,
			Type: "aws_cloudfront_distribution",
			Spec: map[string]any{
				"comment": fmt.Sprintf("%s %s %s", config.Project, config.Environment, res.Name),
			},
		})
	case ResourceServerless:
		resources = append(resources, Resource{
			Name: res.Name,
			Type: "aws_lambda_function",
			Spec: map[string]any{
				"function_name": fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name),
				"runtime":       "python3.12",
			},
		})
	case ResourceMonitoring:
		resources = append(resources, Resource{
			Name: res.Name,
			Type: "aws_cloudwatch_metric_alarm",
			Spec: map[string]any{
				"alarm_name": fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name),
			},
		})
	case ResourceSecrets:
		resources = append(resources, Resource{
			Name: res.Name,
			Type: "aws_secretsmanager_secret",
			Spec: map[string]any{
				"name": fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name),
			},
		})
	case ResourceDNS:
		resources = append(resources, Resource{
			Name: res.Name,
			Type: "aws_route53_zone",
			Spec: map[string]any{
				"name": res.Name,
			},
		})
	case ResourceNetworking:
		resources = append(resources, Resource{
			Name: res.Name,
			Type: "aws_vpc",
			Spec: map[string]any{
				"cidr_block": "10.0.0.0/16",
			},
		})
	}
}

	estimator := NewCostEstimator()
	cost := estimator.EstimateCost(string(config.Provider), resources)

	return &PlanResult{
		Resources:    resources,
		CostEstimate: cost,
	}, nil
}

func (a *AWSAdapter) Deploy(config *DeployConfig) (*DeployResult, error) {
	if err := a.Validate(config); err != nil {
		return nil, err
	}

	planResult, err := a.Plan(config)
	if err != nil {
		return nil, err
	}

	tf, err := a.ExportTerraform(config)
	if err != nil {
		return nil, err
	}

	pool := GetDefaultPool()
	tr, pooled := pool.Get(config.Project, AWS)
	if pooled {
		if err := tr.writeHCL(tf); err != nil {
			return nil, err
		}
		if err := tr.Apply(); err != nil {
			return nil, err
		}
	} else {
		workDir, err := TempWorkDir("naeos-aws")
		if err != nil {
			return nil, err
		}
		tr = NewTerraformRunner(workDir)
		if a.Runner != nil {
			tr.Runner = a.Runner
		}
		if err := tr.Deploy(tf); err != nil {
			return nil, err
		}
		pool.Put(config.Project, AWS, tr, true)
	}

	deployed := []DeployedResource{}
	for _, res := range planResult.Resources {
		deployed = append(deployed, DeployedResource{
			Name: res.Name,
			Type: res.Type,
			ID:   fmt.Sprintf("arn:aws:%s:%s:%s", res.Type, config.Region, res.Name),
		})
	}

	result := &DeployResult{
		Provider:  AWS,
		Resources: deployed,
		Terraform: tf,
		Status:    "deployed",
		Timestamp: time.Now(),
	}

	sm := NewStateManager()
	sm.Save(&DeploymentRecord{
		Project:     config.Project,
		Provider:    AWS,
		Environment: config.Environment,
		Region:      config.Region,
		Resources:   deployed,
		TerraformDir: tr.WorkDir,
		Timestamp:   result.Timestamp,
		Status:      "deployed",
	})

	return result, nil
}

func (a *AWSAdapter) Destroy(config *DeployConfig) error {
	pool := GetDefaultPool()
	if tr, pooled := pool.Get(config.Project, AWS); pooled {
		if err := tr.ApplyDestroy(); err == nil {
			pool.Remove(config.Project, AWS)
			sm := NewStateManager()
			sm.Delete(config.Project, AWS)
			return nil
		}
	}

	sm := NewStateManager()
	record, err := sm.Load(config.Project, AWS)
	if err == nil && record.TerraformDir != "" {
		tr := NewTerraformRunner(record.TerraformDir)
		if a.Runner != nil {
			tr.Runner = a.Runner
		}
		if derr := tr.DestroyAll(); derr == nil {
			sm.Delete(config.Project, AWS)
			return nil
		}
	}

	planResult, err := a.Plan(config)
	if err != nil {
		return err
	}
	if len(planResult.Resources) == 0 {
		return fmt.Errorf("no resources to destroy")
	}

	tf, err := a.ExportTerraform(config)
	if err != nil {
		return err
	}

	workDir, werr := TempWorkDir("naeos-aws-destroy")
	if werr != nil {
		return werr
	}

	tr := NewTerraformRunner(workDir)
	if a.Runner != nil {
		tr.Runner = a.Runner
	}
	if err := tr.writeHCL(tf); err != nil {
		return err
	}
	if err := tr.DestroyAll(); err != nil {
		return err
	}

	sm.Delete(config.Project, AWS)
	return nil
}

func (a *AWSAdapter) ExportTerraform(config *DeployConfig) (string, error) {
	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf(`terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = "%s"
}

`, config.Region))

	// Local variables for naming
	sb.WriteString(fmt.Sprintf(`locals {
  project     = "%s"
  environment = "%s"
  common_tags = {
    Environment = "%s"
    Project     = "%s"
    ManagedBy   = "%s"
  }
}

`, config.Project, config.Environment, config.Environment, config.Project, version.ProductName))

	for _, res := range config.Resources {
		switch res.Type {
		case ResourceStorage:
			bucketName := fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name)
			sb.WriteString(fmt.Sprintf(`resource "aws_s3_bucket" "%s" {
  bucket = "%s"

  tags = local.common_tags
}

resource "aws_s3_bucket_versioning" "%s" {
  bucket = aws_s3_bucket.%s.id

  versioning_configuration {
    status = "Enabled"
  }
}

`, res.Name, bucketName, res.Name, res.Name))

		case ResourceCompute:
			clusterName := fmt.Sprintf("%s-%s", config.Project, config.Environment)
			sb.WriteString(fmt.Sprintf(`resource "aws_ecs_cluster" "%s" {
  name = "%s"

  setting {
    name  = "containerInsights"
    value = "enabled"
  }

  tags = local.common_tags
}

resource "aws_iam_role" "%s_execution" {
  name = "%s-%s-execution"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ecs-tasks.amazonaws.com"
      }
    }]
  })

  tags = local.common_tags
}

resource "aws_ecs_task_definition" "%s" {
  family                   = "%s-%s"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = 256
  memory                   = 512

  execution_role_arn = aws_iam_role.%s_execution.arn

  container_definitions = jsonencode([{
    name  = "%s"
    image = "%s:latest"
    portMappings = [{
      containerPort = 8080
      hostPort      = 8080
    }]
  }])

  tags = local.common_tags
}

resource "aws_ecs_service" "%s" {
  name            = "%s"
  cluster         = aws_ecs_cluster.%s.id
  task_definition = aws_ecs_task_definition.%s.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = []
    security_groups  = []
    assign_public_ip = true
  }

  tags = local.common_tags
}

`, res.Name, clusterName,
			res.Name, config.Project, res.Name,
			res.Name, config.Project, res.Name,
			res.Name, res.Name,
			res.Name, res.Name,
			res.Name, res.Name, res.Name))

		case ResourceDatabase:
			identifier := fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name)
			sb.WriteString(fmt.Sprintf(`resource "random_password" "%s_db" {
  length           = 32
  special          = true
  override_special = "!#$%%^&*()-_=+[]{}<>:?"
}

resource "aws_db_subnet_group" "%s" {
  name       = "%s-subnet"
  subnet_ids = []

  tags = local.common_tags
}

resource "aws_security_group" "%s" {
  name        = "%s-sg"
  description = "Security group for %s"
  vpc_id      = ""

  tags = local.common_tags
}

resource "aws_rds_instance" "%s" {
  identifier     = "%s"
  engine         = "postgres"
  engine_version = "15"
  instance_class = "db.t3.micro"

  allocated_storage     = 20
  max_allocated_storage = 100
  storage_encrypted     = true

  db_name  = "%s"
  username = "admin"
  password = random_password.%s_db.result

  db_subnet_group_name   = aws_db_subnet_group.%s.name
  vpc_security_group_ids = [aws_security_group.%s.id]

  backup_retention_period = 7
  multi_az               = false
  skip_final_snapshot    = true

  tags = local.common_tags
}

`, res.Name,
			res.Name, identifier,
			res.Name, identifier, res.Name,
			res.Name, identifier, res.Name,
			res.Name, res.Name, res.Name))

		case ResourceCache:
			clusterID := fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name)
			sb.WriteString(fmt.Sprintf(`resource "aws_elasticache_subnet_group" "%s" {
  name       = "%s-subnet"
  subnet_ids = []

  tags = local.common_tags
}

resource "aws_elasticache_cluster" "%s" {
  cluster_id           = "%s"
  engine               = "redis"
  engine_version       = "7.0"
  node_type            = "cache.t3.micro"
  num_cache_nodes      = 1
  parameter_group_name = "default.redis7"
  port                 = 6379

  subnet_group_name  = aws_elasticache_subnet_group.%s.name
  security_group_ids = []

  tags = local.common_tags
}

`, res.Name, res.Name,
			res.Name, clusterID,
			res.Name))

		case ResourceQueue:
			queueName := fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name)
			sb.WriteString(fmt.Sprintf(`resource "aws_sqs_queue" "%s" {
  name                      = "%s"
  delay_seconds             = 0
  max_message_size          = 262144
  message_retention_seconds = 345600
  receive_wait_time_seconds = 10

  tags = local.common_tags
}

resource "aws_sqs_queue_policy" "%s_policy" {
  queue_url = aws_sqs_queue.%s.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Sid       = "%sAllowAll"
      Effect    = "Allow"
      Principal = "*"
      Action    = "sqs:*"
      Resource  = aws_sqs_queue.%s.arn
    }]
  })
}

`, res.Name, queueName,
			res.Name, res.Name,
			res.Name, res.Name))

		case ResourceCDN:
			sb.WriteString(fmt.Sprintf(`resource "aws_cloudfront_distribution" "%s" {
  comment = "%s"
  enabled = true

  default_cache_behavior {
    allowed_methods  = ["GET", "HEAD", "OPTIONS"]
    cached_methods   = ["GET", "HEAD"]
    target_origin_id = "origin"

    forwarded_values {
      query_string = false
      cookies {
        forward = "none"
      }
    }

    viewer_protocol_policy = "redirect-to-https"
    min_ttl                = 0
    default_ttl            = 3600
    max_ttl                = 86400
  }

  origin {
    domain_name = "example.com"
    origin_id   = "origin"
  }

  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }

  tags = local.common_tags
}

			`, res.Name, fmt.Sprintf("%s %s %s", config.Project, config.Environment, res.Name)))

		case ResourceServerless:
			funcName := fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name)
			sb.WriteString(fmt.Sprintf(`resource "aws_lambda_function" "%s" {
  function_name = "%s"
  runtime       = "python3.12"
  handler       = "index.handler"
  role          = aws_iam_role.%s_lambda.arn
  memory_size   = 256
  timeout       = 30

  environment {
    variables = {
      ENVIRONMENT = "%s"
    }
  }

  tags = local.common_tags
}

resource "aws_iam_role" "%s_lambda" {
  name = "%s-lambda"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "lambda.amazonaws.com"
      }
    }]
  })

  tags = local.common_tags
}

`, res.Name, funcName,
			res.Name, config.Environment,
			res.Name, res.Name))

		case ResourceMonitoring:
			alarmName := fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name)
			sb.WriteString(fmt.Sprintf(`resource "aws_cloudwatch_metric_alarm" "%s" {
  alarm_name          = "%s"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "CPUUtilization"
  namespace           = "AWS/EC2"
  period              = 300
  statistic           = "Average"
  threshold           = 80
  alarm_description   = "%s %s %s alarm"

  tags = local.common_tags
}

`, res.Name, alarmName,
			config.Project, config.Environment, res.Name))

		case ResourceSecrets:
			secretName := fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name)
			sb.WriteString(fmt.Sprintf(`resource "aws_secretsmanager_secret" "%s" {
  name = "%s"

  tags = local.common_tags
}

`, res.Name, secretName))

		case ResourceDNS:
			sb.WriteString(fmt.Sprintf(`resource "aws_route53_zone" "%s" {
  name = "%s"

  tags = local.common_tags
}

`, res.Name, res.Name))

		case ResourceNetworking:
			sb.WriteString(fmt.Sprintf(`resource "aws_vpc" "%s" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = local.common_tags
}

resource "aws_subnet" "%s" {
  vpc_id            = aws_vpc.%s.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = "%sa"

  tags = local.common_tags
}

`, res.Name, res.Name,
			res.Name, res.Name))

		}
	}

	return sb.String(), nil
}
