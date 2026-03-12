import React, { useEffect, useState } from "react";

const API_URL = import.meta.env.VITE_API_URL || "";
const SQLITE_API_URL = import.meta.env.VITE_SQLITE_API_URL || "";
const STATIC_JSON_URL = import.meta.env.VITE_STATIC_JSON_URL || "";

export default function App() {
    const [users, setUsers] = useState([]);
    const [notes, setNotes] = useState([]);
    const [config, setConfig] = useState(null);
    const [products, setProducts] = useState([]);

    useEffect(() => {
        fetch(`${API_URL}/api/users`).then((r) => r.json()).then(setUsers).catch(console.error);
        fetch(`${SQLITE_API_URL}/api/notes`).then((r) => r.json()).then(setNotes).catch(console.error);
        fetch(`${STATIC_JSON_URL}/api/config.json`).then((r) => r.json()).then(setConfig).catch(console.error);
        fetch(`${STATIC_JSON_URL}/api/products.json`).then((r) => r.json()).then(setProducts).catch(console.error);
    }, []);

    return (
        <div style={{ fontFamily: "system-ui", maxWidth: 800, margin: "0 auto", padding: 20 }}>
            <h1>Fullstack Preview Dashboard</h1>

            <section>
                <h2>Users (PostgreSQL via API)</h2>
                <ul>{users.map((u) => <li key={u.id}>{u.name} — {u.email}</li>)}</ul>
            </section>

            <section>
                <h2>Notes (SQLite via SQLite-API)</h2>
                <ul>{notes.map((n) => <li key={n.id}>{n.title}: {n.content}</li>)}</ul>
            </section>

            <section>
                <h2>Products (Static JSON)</h2>
                <ul>{products.map((p) => <li key={p.id}>{p.name} — ${p.price}</li>)}</ul>
            </section>

            <section>
                <h2>Config (Static JSON)</h2>
                <pre>{JSON.stringify(config, null, 2)}</pre>
            </section>
        </div>
    );
}
