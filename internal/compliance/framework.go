package compliance

import (
	"fmt"
	"time"

	"github.com/NAEOS-foundation/naeos/internal/audit"
)

type Framework string

const (
	FrameworkSOC2  Framework = "soc2"
	FrameworkHIPAA Framework = "hipaa"
	FrameworkGDPR  Framework = "gdpr"
)

type Control struct {
	ID          string
	Title       string
	Description string
	Category    string
}

type FrameworkDef struct {
	Name        string
	Version     string
	Description string
	Controls    []Control
}

var Frameworks = map[Framework]*FrameworkDef{
	FrameworkSOC2: {
		Name:        "SOC 2 Type II",
		Version:     "2024",
		Description: "Service Organization Control 2 — security, availability, processing integrity, confidentiality, privacy",
		Controls: []Control{
			{ID: "CC1.1", Title: "Control Environment", Description: "Board and management oversight", Category: "Security"},
			{ID: "CC2.1", Title: "Communication and Information", Description: "Relevant information identified and communicated", Category: "Security"},
			{ID: "CC3.1", Title: "Risk Assessment", Description: "Identifies and assesses risks", Category: "Risk"},
			{ID: "CC4.1", Title: "Monitoring Activities", Description: "Ongoing monitoring of controls", Category: "Monitoring"},
			{ID: "CC5.1", Title: "Control Activities", Description: "Policies and procedures established", Category: "Security"},
			{ID: "CC6.1", Title: "Logical and Physical Access", Description: "Access restricted to authorized users", Category: "Access"},
			{ID: "CC7.1", Title: "System Operations", Description: "System operations monitored", Category: "Operations"},
			{ID: "CC8.1", Title: "Change Management", Description: "Changes authorized and tested", Category: "Change"},
		},
	},
	FrameworkHIPAA: {
		Name:        "HIPAA Security Rule",
		Version:     "2024",
		Description: "Health Insurance Portability and Accountability Act — administrative, physical, technical safeguards",
		Controls: []Control{
			{ID: "164.308(a)(1)", Title: "Security Management Process", Description: "Risk analysis and management", Category: "Administrative"},
			{ID: "164.308(a)(3)", Title: "Workforce Security", Description: "Authorized access to ePHI", Category: "Administrative"},
			{ID: "164.308(a)(4)", Title: "Information Access Management", Description: "Access authorization", Category: "Administrative"},
			{ID: "164.308(a)(5)", Title: "Security Awareness and Training", Description: "Workforce training", Category: "Administrative"},
			{ID: "164.308(a)(6)", Title: "Security Incident Procedures", Description: "Incident response", Category: "Administrative"},
			{ID: "164.308(a)(7)", Title: "Contingency Plan", Description: "Data backup and disaster recovery", Category: "Administrative"},
			{ID: "164.308(a)(8)", Title: "Evaluation", Description: "Periodic technical evaluation", Category: "Administrative"},
			{ID: "164.310(a)(1)", Title: "Facility Access Controls", Description: "Physical access controls", Category: "Physical"},
			{ID: "164.312(a)(1)", Title: "Access Control", Description: "Unique user identification", Category: "Technical"},
			{ID: "164.312(c)(1)", Title: "Person or Entity Authentication", Description: "Authentication mechanisms", Category: "Technical"},
			{ID: "164.312(d)", Title: "Transmission Security", Description: "Encryption of ePHI in transit", Category: "Technical"},
		},
	},
	FrameworkGDPR: {
		Name:        "GDPR",
		Version:     "2018",
		Description: "General Data Protection Regulation — data protection, privacy, consent, breach notification",
		Controls: []Control{
			{ID: "Art. 5", Title: "Principles Relating to Processing", Description: "Lawfulness, fairness, transparency", Category: "Principles"},
			{ID: "Art. 7", Title: "Consent", Description: "Explicit consent for data processing", Category: "Consent"},
			{ID: "Art. 17", Title: "Right to Erasure", Description: "Right to be forgotten", Category: "Data Rights"},
			{ID: "Art. 20", Title: "Data Portability", Description: "Right to data portability", Category: "Data Rights"},
			{ID: "Art. 32", Title: "Security of Processing", Description: "Appropriate technical measures", Category: "Security"},
			{ID: "Art. 33", Title: "Data Breach Notification", Description: "72-hour breach notification", Category: "Breach"},
			{ID: "Art. 35", Title: "Data Protection Impact Assessment", Description: "DPIA for high-risk processing", Category: "Assessment"},
			{ID: "Art. 37", Title: "Data Protection Officer", Description: "DPO designation", Category: "Governance"},
		},
	},
}

type ComplianceReport struct {
	Framework       Framework
	GeneratedAt     time.Time
	TotalControls   int
	PassedControls  int
	FailedControls  int
	NotApplicable   int
	ControlStatuses []ControlStatus
	Evidence        []Evidence
}

type ControlStatus struct {
	ControlID string
	Passed    bool
	Finding   string
	Evidence  []string
}

type Evidence struct {
	Type        string
	Description string
	Source      string
	Timestamp   time.Time
}

func GenerateReport(framework Framework, events []audit.AuditEvent) *ComplianceReport {
	def, ok := Frameworks[framework]
	if !ok {
		return nil
	}

	report := &ComplianceReport{
		Framework:       framework,
		GeneratedAt:     time.Now(),
		TotalControls:   len(def.Controls),
		ControlStatuses: make([]ControlStatus, 0, len(def.Controls)),
	}

	for _, c := range def.Controls {
		status := evaluateControl(c, events)
		report.ControlStatuses = append(report.ControlStatuses, status)
		if status.Passed {
			report.PassedControls++
		} else {
			report.FailedControls++
		}

		if status.ControlID != "" {
			report.Evidence = append(report.Evidence, Evidence{
				Type:        "audit-log",
				Description: fmt.Sprintf("Control %s: %s", c.ID, c.Title),
				Source:      "naeos audit log",
				Timestamp:   time.Now(),
			})
		}
	}

	return report
}

func evaluateControl(c Control, events []audit.AuditEvent) ControlStatus {
	switch c.ID {
	case "CC6.1", "164.312(a)(1)", "Art. 32":
		for _, e := range events {
			if e.Resource == "user" && e.Action == "create" {
				return ControlStatus{ControlID: c.ID, Passed: true, Finding: "User access controls are in place"}
			}
		}
		return ControlStatus{ControlID: c.ID, Passed: false, Finding: "No user access events found — implement access controls"}
	case "CC7.1", "164.308(a)(6)":
		for _, e := range events {
			if e.Status == "failed" || e.Status == "error" {
				return ControlStatus{ControlID: c.ID, Passed: true, Finding: "System operations monitored with incident detection"}
			}
		}
		return ControlStatus{ControlID: c.ID, Passed: false, Finding: "No incident monitoring detected"}
	case "CC4.1", "164.308(a)(8)":
		if len(events) > 0 {
			return ControlStatus{ControlID: c.ID, Passed: true, Finding: "Audit trail present with monitoring activities"}
		}
		return ControlStatus{ControlID: c.ID, Passed: false, Finding: "No audit events detected"}
	case "Art. 17":
		for _, e := range events {
			if e.Action == "delete" && e.Resource == "user" {
				return ControlStatus{ControlID: c.ID, Passed: true, Finding: "User deletion (right to erasure) supported"}
			}
		}
		return ControlStatus{ControlID: c.ID, Passed: false, Finding: "No user deletion events — verify right to erasure"}
	default:
		if len(events) > 0 {
			return ControlStatus{ControlID: c.ID, Passed: true, Finding: "General compliance evidence available"}
		}
		return ControlStatus{ControlID: c.ID, Passed: false, Finding: "No evidence collected"}
	}
}
