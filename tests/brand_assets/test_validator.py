# Copyright 2025 NAEOS contributors
# SPDX-License-Identifier: Apache-2.0

import importlib.util
import tempfile
import unittest
from pathlib import Path

SCRIPT = Path(__file__).resolve().parents[2] / "scripts" / "validate-brand-assets.py"
SPEC = importlib.util.spec_from_file_location("brand_validator", SCRIPT)
validator = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(validator)

class SVGValidationTests(unittest.TestCase):
    def write_svg(self, text):
        temp = tempfile.NamedTemporaryFile("w", suffix=".svg", encoding="utf-8", delete=False)
        temp.write(text)
        temp.close()
        self.addCleanup(Path(temp.name).unlink, missing_ok=True)
        return Path(temp.name)

    def test_valid_svg_passes(self):
        path = self.write_svg('<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10"><rect width="10" height="10" fill="#fff"/></svg>')
        self.assertEqual([], validator.svg_validate(path))

    def test_external_resource_is_rejected(self):
        path = self.write_svg('<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10"><image href="https://example.invalid/image.png"/></svg>')
        self.assertTrue(any("external resource" in issue for issue in validator.svg_validate(path)))

    def test_file_and_root_relative_resources_are_rejected(self):
        for reference in ("file:///etc/passwd", "/assets/remote.svg"):
            path = self.write_svg(f'<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10"><image href="{reference}"/></svg>')
            self.assertTrue(any("external resource" in issue for issue in validator.svg_validate(path)))

    def test_script_and_event_attribute_rejected(self):
        path = self.write_svg('<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10" onload="alert(1)"><script>alert(1)</script></svg>')
        issues = validator.svg_validate(path)
        self.assertTrue(any("script" in issue for issue in issues))
        self.assertTrue(any("event-handler" in issue for issue in issues))

    def test_missing_viewbox_rejected(self):
        path = self.write_svg('<svg xmlns="http://www.w3.org/2000/svg"><rect width="1" height="1"/></svg>')
        self.assertTrue(any("viewBox" in issue for issue in validator.svg_validate(path)))

    def test_malformed_xml_rejected(self):
        path = self.write_svg("<svg><")
        self.assertTrue(any("parse error" in issue for issue in validator.svg_validate(path)))

    def test_external_css_in_style_element_is_rejected(self):
        path = self.write_svg('<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10"><style>.x { fill: url(https://example.invalid/a.svg); }</style></svg>')
        self.assertTrue(any("external CSS resource" in issue for issue in validator.svg_validate(path)))

    def test_viewbox_accepts_whitespace_separators(self):
        path = self.write_svg('<svg xmlns="http://www.w3.org/2000/svg" viewBox="0\n 0\t10 10"><rect width="10" height="10"/></svg>')
        self.assertEqual([], validator.svg_validate(path))

if __name__ == "__main__":
    unittest.main()
