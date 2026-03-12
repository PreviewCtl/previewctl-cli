import { defineConfig } from "astro/config";
import starlight from "@astrojs/starlight";

export default defineConfig({
    site: process.env.SITE_URL || "",
    base: process.env.BASE_PATH || "",
    integrations: [
        starlight({
            title: "PreviewCTL",
            logo: {
                src: "./src/assets/logo.png",
            },
            favicon: "./src/assets/logo.png",
            description:
                "Spin up ephemeral preview environments with Docker — one YAML, one command.",
            social: [
                {
                    icon: "github",
                    label: "GitHub",
                    href: "https://github.com/previewctl/previewctl-cli",
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
                        { label: "validate", slug: "commands/validate" },
                        { label: "up", slug: "commands/up" },
                        { label: "down", slug: "commands/down" },
                        { label: "list", slug: "commands/list" },
                        { label: "delete", slug: "commands/delete" },
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
