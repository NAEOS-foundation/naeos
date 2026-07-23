CREATE TABLE "newsletter_subscribers" (
	"id" serial PRIMARY KEY,
	"email" text NOT NULL UNIQUE,
	"locale" text DEFAULT 'en' NOT NULL,
	"source" text DEFAULT 'footer' NOT NULL,
	"created_at" timestamp with time zone DEFAULT now() NOT NULL
);
