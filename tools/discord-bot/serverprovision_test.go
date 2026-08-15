package main

import "testing"

func TestDefaultPlanStructure(t *testing.T) {
	plan := defaultPlan()

	if len(plan.Roles) == 0 {
		t.Fatal("expected at least one role")
	}
	for _, r := range plan.Roles {
		if r.Name == "" {
			t.Fatal("role without name")
		}
	}

	seen := make(map[string]bool)
	seenCat := make(map[string]bool)
	for _, cat := range plan.Categories {
		if seenCat[cat.Name] {
			t.Fatalf("duplicate category %q", cat.Name)
		}
		seenCat[cat.Name] = true
		if len(cat.Channels) == 0 {
			t.Fatalf("category %q has no channels", cat.Name)
		}
		for _, ch := range cat.Channels {
			if ch.Name == "" {
				t.Fatalf("channel without name in category %q", cat.Name)
			}
			if seen[ch.Name] {
				t.Fatalf("duplicate channel %q", ch.Name)
			}
			seen[ch.Name] = true
		}
	}

	// Blueprint expectations: staff category is private, announcements locked.
	foundPrivate := false
	foundLocked := false
	for _, cat := range plan.Categories {
		if cat.Name == "MODERATION" {
			for _, ch := range cat.Channels {
				if ch.Private {
					foundPrivate = true
				}
				if ch.Locked {
					foundLocked = true
				}
			}
		}
	}
	if !foundPrivate || !foundLocked {
		t.Fatal("MODERATION category should contain private and locked channels")
	}
}
