const express = require("express");
const { Pool } = require("pg");

const app = express();
app.use(express.json());

const pool = new Pool({ connectionString: process.env.DATABASE_URL });

// Health check
app.get("/health", async (_req, res) => {
    try {
        await pool.query("SELECT 1");
        res.json({ status: "ok", preview: process.env.PREVIEW_ID });
    } catch (err) {
        res.status(500).json({ status: "error", message: err.message });
    }
});

// List users from postgres
app.get("/api/users", async (_req, res) => {
    const { rows } = await pool.query("SELECT id, name, email FROM users ORDER BY id");
    res.json(rows);
});

// Create user
app.post("/api/users", async (req, res) => {
    const { name, email } = req.body;
    const { rows } = await pool.query(
        "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id, name, email",
        [name, email]
    );
    res.status(201).json(rows[0]);
});

// Preview info
app.get("/api/info", (_req, res) => {
    res.json({
        app: process.env.APP_LABEL,
        preview_id: process.env.PREVIEW_ID,
        environment: process.env.NODE_ENV,
    });
});

app.listen(8080, () => console.log(`API listening on :${8080}`));
