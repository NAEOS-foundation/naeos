package cloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/version"
)

type AWSAdapter struct{}

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

func (a *AWSAdapter) Plan(config *DeployConfig) ([]Resource, error) {
	resources := []Resource{}
	for _, res := range config.Resources {
		switch res.Type {
		case ResourceStorage:
			resources = append(resources, Resource{
				Name: res.Name,
				Type: "aws_s3_bucket",
				Spec: map[string]interface{}{
					"bucket": fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name),
					"region": config.Region,
				},
			})
		case ResourceCompute:
			resources = append(resources, Resource{
				Name: res.Name,
				Type: "aws_ecs_service",
				Spec: map[string]interface{}{
					"cluster": fmt.Sprintf("%s-%s", config.Project, config.Environment),
					"service": res.Name,
				},
			})
		case ResourceDatabase:
			resources = append(resources, Resource{
				Name: res.Name,
				Type: "aws_rds_instance",
				Spec: map[string]interface{}{
					"identifier": fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name),
					"engine":     "postgres",
				},
			})
		case ResourceCache:
			resources = append(resources, Resource{
				Name: res.Name,
				Type: "aws_elasticache_cluster",
				Spec: map[string]interface{}{
					"cluster_id": fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name),
					"engine":     "redis",
				},
			})
		case ResourceQueue:
			resources = append(resources, Resource{
				Name: res.Name,
				Type: "aws_sqs_queue",
				Spec: map[string]interface{}{
					"name": fmt.Sprintf("%s-%s-%s", config.Project, config.Environment, res.Name),
				},
			})
		case ResourceCDN:
			resources = append(resources, Resource{
				Name: res.Name,
				Type: "aws_cloudfront_distribution",
				Spec: map[string]interface{}{
					"comment": fmt.Sprintf("%s %s %s", config.Project, config.Environment, res.Name),
				},
			})
		}
	}
	return resources, nil
}

func (a *AWSAdapter) Deploy(config *DeployConfig) (*DeployResult, error) {
	plan, err := a.Plan(config)
	if err != nil {
		return nil, err
	}

	deployed := []DeployedResource{}
	for _, res := range plan {
		deployed = append(deployed, DeployedResource{
			Name: res.Name,
			Type: res.Type,
			ID:   fmt.Sprintf("arn:aws:%s:%s:%s", res.Type, config.Region, res.Name),
		})
	}

	tf, _ := a.ExportTerraform(config)

	return &DeployResult{
		Provider:  AWS,
		Resources: deployed,
		Terraform: tf,
		Status:    "deployed",
		Timestamp: time.Now(),
	}, nil
}

func (a *AWSAdapter) Destroy(config *DeployConfig) error {
	plan, err := a.Plan(config)
	if err != nil {
		return err
	}
	if len(plan) == 0 {
		return fmt.Errorf("no resources to destroy")
	}
	var resources []string
	for _, r := range plan {
		resources = append(resources, r.Type+"."+r.Name)
	}
	return fmt.Errorf("terraform destroy planned: %d AWS resources would be destroyed: %s", len(resources), strings.Join(resources, ", "))
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
			sb.WriteString(fmt.Sprintf(`resource "aws_db_subnet_group" "%s" {
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
  password = ""

  db_subnet_group_name   = aws_db_subnet_group.%s.name
  vpc_security_group_ids = [aws_security_group.%s.id]

  backup_retention_period = 7
  multi_az               = false
  skip_final_snapshot    = true

  tags = local.common_tags
}

`, res.Name, identifier,
			res.Name, identifier, res.Name,
			res.Name, identifier, res.Name,
			res.Name, res.Name))

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
		}
	}

	return sb.String(), nil
}
