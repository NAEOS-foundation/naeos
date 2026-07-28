import sharp from "sharp";
import fs from "fs";

await Promise.all([
  sharp(fs.readFileSync("static/images/og-dark.svg"))
    .resize(1200, 630, { fit: "fill" })
    .png()
    .toFile("static/images/og-default.png"),
  sharp(fs.readFileSync("static/images/og-light.svg"))
    .resize(1200, 630, { fit: "fill" })
    .png()
    .toFile("static/images/og-default-light.png"),
]);

console.log("OG images generated");
