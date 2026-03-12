import os
import sqlite3
from flask import Flask, jsonify, request

app = Flask(__name__)
DB_PATH = os.environ.get("DB_PATH", "/app/data/app.db")


def get_db():
    conn = sqlite3.connect(DB_PATH)
    conn.row_factory = sqlite3.Row
    return conn


@app.route("/health")
def health():
    return jsonify(status="ok", preview=os.environ.get("PREVIEW_ID"))


@app.route("/api/notes")
def list_notes():
    db = get_db()
    rows = db.execute("SELECT id, title, content FROM notes ORDER BY id").fetchall()
    db.close()
    return jsonify([dict(r) for r in rows])


@app.route("/api/notes", methods=["POST"])
def create_note():
    data = request.get_json()
    db = get_db()
    cursor = db.execute(
        "INSERT INTO notes (title, content) VALUES (?, ?)",
        (data["title"], data["content"]),
    )
    db.commit()
    note_id = cursor.lastrowid
    db.close()
    return jsonify(id=note_id, title=data["title"], content=data["content"]), 201


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=int(os.environ.get("PORT", 5000)))
