import { defineConfig } from "astro/config";
import starlight from "@astrojs/starlight";

export default defineConfig({
    site: "https://previewctl.dev",
    integrations: [
        starlight({
            title: "PreviewCtrl",
            description:
                "Spin up ephemeral preview environments with Docker — one YAML, one command.",
            social: [
                {
                    icon: "github",
                    label: "GitHub",
                    href: "https://github.com/previewctrl/previewctl-cli",
                },
            ],
            customCss: ["./src/styles/landing.css"],
            sidebar: [
                {
                    label: "Getting Started",
                    items: [{ label: "Quick Start", slug: "getting-started" }],
                },
                {
                    label: "Commands",
                    items: [
                        { label: "init", slug: "commands/init" },
                        { label: "up", slug: "commands/up" },
                        { label: "list", slug: "commands/list" },
                        { label: "delete", slug: "commands/delete" },
                        { label: "validate", slug: "commands/validate" },
                    ],
                },
                {
                    label: "Reference",
                    items: [{ label: "Configuration", slug: "configuration" }],
                },
            ],
        }),
    ],
});
