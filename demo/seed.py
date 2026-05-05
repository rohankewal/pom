#!/usr/bin/env python3
"""Seeds ~/.go-pom/sessions.json with realistic demo data."""
import json, datetime, random, os, pathlib

random.seed(42)

titles = ["deep work", "api redesign", "bug fixes", "code review", "feature work", "writing"]
tag_sets = [["code"], ["code", "focus"], ["review"], ["writing"], ["work", "focus"]]

sessions = []
today = datetime.date.today()

for days_ago in range(1, 22):
    d = today - datetime.timedelta(days=days_ago)
    # fewer sessions on weekends
    if d.weekday() >= 5:
        n = random.choice([0, 0, 1])
    else:
        n = random.choice([1, 2, 2, 3])

    for j in range(n):
        rounds = random.choice([4, 4, 4, 3, 2])
        completed = rounds if random.random() > 0.15 else rounds - 1
        hour = 9 + j * 3
        sessions.append({
            "date": f"{d.isoformat()}T{hour:02d}:00:00Z",
            "title": random.choice(titles),
            "tags": random.choice(tag_sets),
            "rounds_planned": rounds,
            "rounds_completed": completed,
            "work_duration": 1500000000000,  # 25 minutes in nanoseconds
            "interruptions": random.randint(0, 2),
            "completed": completed == rounds,
            "note": "",
        })

sessions.sort(key=lambda s: s["date"])

path = pathlib.Path(os.path.expanduser("~/.go-pom"))
path.mkdir(exist_ok=True)
out = path / "sessions.json"
json.dump(sessions, open(out, "w"), indent=2)
print(f"Seeded {len(sessions)} sessions into {out}")
