const canvas = document.getElementById("city");

const ctx = canvas.getContext("2d");
const status = document.getElementById("status");


function draw(world) {
    ctx.clearRect(0, 0, canvas.width, canvas.height);

    const scale = 3;
    const offset = 70;

    for (const road of Object.values(world.roads)) {
        // Рисуем очередную дорогу.
        ctx.beginPath();

        for (let i = 0; i < road.Points.length; i++) {
            const point = road.Points[i];
            const x = offset + point.X * scale;
            const y = offset + point.Y * scale;

            if (i === 0) {
                ctx.moveTo(x, y);
            } else {
                ctx.lineTo(x, y);
            }
        }

        ctx.strokeStyle = "#444";
        ctx.lineWidth = 40;
        ctx.lineJoin = "round";
        ctx.stroke();

    }

    // Каждый светофор привязан к своей дороге.
    for (const light of world.trafficLights ?? []) {
        const road = world.roads[light.RoadID];
        if (!road) continue;
        const point = pointOnRoad(road, light.Position);

        ctx.save();
        ctx.translate(offset + point.x * scale, offset + point.y * scale);
        ctx.rotate(point.angle);
        ctx.fillStyle = "white";
        ctx.fillRect(-1, -20, 2, 40);
        ctx.fillStyle = light.Green ? "green" : "red";
        ctx.fillRect(-6, -40, 12, 14);
        ctx.restore();
    }

    // Рисуем машины на их текущих дорогах.
    for (const car of world.cars ?? []) {
        const roadID = car.Route[car.RoadIndex];
        const road = world.roads[roadID];
        const point = pointOnRoad(road, car.Position);

        ctx.save();
        ctx.translate(
            offset + point.x * scale,
            offset + point.y * scale
        );
        ctx.rotate(point.angle);

        ctx.fillStyle = "#4dabf7";
        ctx.fillRect(-4 * scale, -10, 4 * scale, 20);

        ctx.restore();
    }
}

async function control(action) {
    const message = document.getElementById("command-status");

    try {
        const response = await fetch("/control", {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify({ action })
        });

        if (!response.ok) {
            throw new Error(await response.text());
        }

        message.textContent = "Команда выполнена: " + action;
    } catch (error) {
        message.textContent = "Ошибка: " + error.message;
    }
}
function pointOnRoad(road, distance) {
    const poinsts = road.Points;
    for (let i = 1; i < poinsts.length; i++) {
        const start = poinsts[i - 1];
        const end = poinsts[i];

        const dx = end.X - start.X;
        const dy = end.Y - start.Y;
        const length = Math.hypot(dx, dy);

        if (length == 0) {
            continue;
        }

        if (distance <= length) {
            const ratio = distance / length;
            return {
                x: start.X + dx * ratio,
                y: start.Y + dy * ratio,
                angle: Math.atan2(dy, dx)
            };
        }

        distance -= length;
    }

    const lastPoint = poinsts[poinsts.length - 1];
    return {
        x: lastPoint.X,
        y: lastPoint.Y,
        angle: 0
    };
}

async function refresh() {
    try {
        const response = await fetch("/state");

        if (!response.ok) {
            throw new Error("HTTP " + response.status);
        }

        const world = await response.json();
        draw(world);
        status.textContent = "Машин: " + (world.cars ?? []).length;
    } catch (error) {
        status.textContent = "Ошибка: " + error.message;
    } finally {
        setTimeout(refresh, 100);
    }
}





refresh();