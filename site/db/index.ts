import { drizzle } from "drizzle-orm/netlify-db";

import * as schema from "./schema.js";

export const db = drizzle({ schema });
