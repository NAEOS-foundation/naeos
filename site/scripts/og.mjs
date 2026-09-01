import sharp from "sharp";
import fs from "fs";

await Promise.all([
  sharp(fs.readFileSync("public/images/og-dark.svg"))
    .resize(1200, 630, { fit: "fill" })
    .png()
    .toFile("public/images/og-default.png"),
  sharp(fs.readFileSync("public/images/og-light.svg"))
    .resize(1200, 630, { fit: "fill" })
    .png()
    .toFile("public/images/og-default-light.png"),
]);

console.log("OG images generated");
