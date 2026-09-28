const canvas = document.getElementById("city");
const ctx = canvas.getContext("2d");
const connection = document.getElementById("connection");
const status = document.getElementById("status");
const feedback = document.getElementById("command-status");
const carCount = document.getElementById("car-count");
const roadCount = document.getElementById("road-count");
const lightCount = document.getElementById("light-count");
const arrivedCount = document.getElementById("arrived-count");
const tripTime = document.getElementById("trip-time");
const waitTime = document.getElementById("wait-time");
const pauseButton = document.getElementById("pause");
const resumeButton = document.getElementById("resume");
const inspectorBody = document.getElementById("inspector-body");
const inspectorFeedback = document.getElementById("inspector-feedback");
const carLength = 4;
const laneOffset = 2.35;
const intersectionZoneLength = 12;
const carColors = ["#57cfe0", "#f7b964", "#ed897d", "#9bb5f5", "#8cdfb2"];
let latestWorld = null;
let selectedObject = null;

function resizeCanvas() {
    const bounds = canvas.getBoundingClientRect();
    if (!bounds.width || !bounds.height) return;
    const ratio = Math.min(window.devicePixelRatio || 1, 2);
    canvas.width = Math.round(bounds.width * ratio);
    canvas.height = Math.round(bounds.height * ratio);
    ctx.setTransform(ratio, 0, 0, ratio, 0, 0);
    draw(latestWorld);
}

function draw(world) {
    const width = canvas.clientWidth;
    const height = canvas.clientHeight;
    if (!width || !height) return;
    drawGround(width, height);
    if (!world || !world.roads) return;

    const view = createView(world.roads, world.buildings ?? [], width, height);
    const roadWidth = Math.max(27, view.scale * 9.5);
    const roads = Object.values(world.roads);
    const surfaces = physicalRoads(roads);
    const accessRoads = surfaces.filter(road => road.Kind === "access");
    const groundRoads = surfaces.filter(road => road.Kind !== "access" && !road.Level);
    const elevatedRoads = surfaces.filter(road => road.Level > 0);
    const roundaboutRoads = new Set((world.roundabouts ?? []).flatMap(circle => circle.Segments));

    for (const road of accessRoads) strokeRoad(road, view, Math.max(9, view.scale * 4.5) + 3, "#a1ada6", 0);
    for (const road of accessRoads) strokeRoad(road, view, Math.max(9, view.scale * 4.5), "#687a76", 0);
    for (const road of groundRoads) strokeRoad(road, view, roadWidth + 5, "#a8b5ac", 0);
    for (const road of groundRoads) strokeRoad(road, view, roadWidth, "#3c4b50", 0);

    ctx.save();
    ctx.setLineDash([13, 18]);
    for (const road of groundRoads) {
        if (!roundaboutRoads.has(road.ID)) strokeRoad(road, view, 1.6, "#e6cc8caa", 0);
    }
    ctx.restore();

    for (const circle of world.roundabouts ?? []) drawRoundabout(circle, view, roadWidth);
    for (const junction of world.intersections ?? []) drawJunction(junction, world.roads, view, roadWidth);
    for (const road of surfaces) {
        if (road.OneWay && road.Kind !== "roundabout" && road.Kind !== "access") drawOneWayArrow(road, view);
        const load = physicalRoadLoad(road, world);
        if (load?.waiting > 0 && load.cars >= 2) drawCongestion(road, view, load);
    }
    const cars = world.cars ?? [];
    for (const car of cars) {
        if (!world.roads[car.Route[car.RoadIndex]]?.Level) {
            drawCar(car, world.roads, world.intersections ?? [], world.roundabouts ?? [], view);
        }
    }
    const elevatedWidth = road => road.Kind === "ramp" ? roadWidth * 0.82 : roadWidth;
    ctx.save();
    ctx.globalAlpha = 0.9;
    for (const road of elevatedRoads) strokeRoad(road, view, elevatedWidth(road) + 11, "#a2aaa4", 8);
    for (const road of elevatedRoads) strokeRoad(road, view, elevatedWidth(road) + 6, "#c79c60", 0);
    for (const road of elevatedRoads) strokeRoad(road, view, elevatedWidth(road), "#4d555c", 0);
    ctx.setLineDash([15, 17]);
    for (const road of elevatedRoads) strokeRoad(road, view, 1.8, "#f6d992", 0);
    ctx.restore();
    for (const light of world.trafficLights ?? []) drawTrafficLight(light, world.roads, view, roadWidth);
    for (const building of world.buildings ?? []) drawBuilding(building, world.roads, world.arrivals ?? {}, view);
    drawSelection(world, view);
    for (const road of surfaces) {
        if (road.Closed) drawClosedRoad(road, view);
    }
    for (const car of cars) {
        if (world.roads[car.Route[car.RoadIndex]]?.Level) {
            drawCar(car, world.roads, world.intersections ?? [], world.roundabouts ?? [], view);
        }
    }
}

function drawSelection(world, view) {
    if (!selectedObject) return;
    ctx.save();
    ctx.strokeStyle = "#267d68";
    ctx.lineWidth = selectedObject.kind === "road" ? 7 : 2.5;
    ctx.globalAlpha = 0.85;
    if (selectedObject.kind === "home" || selectedObject.kind === "work") {
        const building = (world.buildings ?? []).find(item => item.ID === selectedObject.id);
        if (building) {
            const point = screenPoint(building.Position, view);
            ctx.beginPath();
            ctx.arc(point.x, point.y, 25, 0, Math.PI * 2);
            ctx.stroke();
        }
    } else {
        const road = world.roads[selectedObject.id];
        if (road && roadPath(road, view)) ctx.stroke();
    }
    ctx.restore();
}

function drawClosedRoad(road, view) {
    if (!roadPath(road, view)) return;
    ctx.save();
    ctx.strokeStyle = "#ff8b76";
    ctx.lineWidth = 5;
    ctx.setLineDash([8, 9]);
    ctx.stroke();
    const middle = pointOnRoad(road, road.Length / 2);
    const point = screenPoint(middle, view);
    ctx.fillStyle = "#402926";
    ctx.strokeStyle = "#ff8b76";
    ctx.setLineDash([]);
    ctx.lineWidth = 2;
    ctx.beginPath();
    ctx.arc(point.x, point.y, 10, 0, Math.PI * 2);
    ctx.fill();
    ctx.stroke();
    ctx.fillStyle = "#ffe9dc";
    ctx.font = "700 13px -apple-system, BlinkMacSystemFont, sans-serif";
    ctx.textAlign = "center";
    ctx.textBaseline = "middle";
    ctx.fillText("×", point.x, point.y);
    ctx.restore();
}

function drawOneWayArrow(road, view) {
    const middle = pointOnRoad(road, road.Length / 2);
    const point = screenPoint({ X: middle.x, Y: middle.y }, view);
    ctx.save();
    ctx.translate(point.x, point.y);
    ctx.rotate(middle.angle);
    ctx.fillStyle = "#f3db9f";
    ctx.beginPath(); ctx.moveTo(8, 0); ctx.lineTo(-5, -5); ctx.lineTo(-5, 5); ctx.closePath(); ctx.fill();
    ctx.restore();
}

function drawCongestion(road, view, load) {
    if (!roadPath(road, view)) return;
    ctx.save();
    ctx.strokeStyle = load.waiting >= 3 ? "#d9675c" : "#d49a5a";
    ctx.lineWidth = 4;
    ctx.globalAlpha = 0.8;
    ctx.stroke();
    ctx.restore();
}

function physicalRoads(roads) {
    const seen = new Set();
    return roads.filter(road => {
        const points = road.Points.map(point => `${point.X},${point.Y}`);
        const forward = points.join("|");
        const reverse = [...points].reverse().join("|");
        const key = forward < reverse ? forward : reverse;
        if (seen.has(key)) return false;
        seen.add(key);
        return true;
    });
}

function physicalRoadLoad(road, world) {
    const points = road.Points;
    const result = { cars: 0, waiting: 0 };
    for (const candidate of Object.values(world.roads ?? {})) {
        if (candidate.Points.length !== points.length) continue;
        const same = candidate.Points.every((point, i) => point.X === points[i].X && point.Y === points[i].Y);
        const reverse = candidate.Points.every((point, i) => point.X === points[points.length - 1 - i].X && point.Y === points[points.length - 1 - i].Y);
        if (!same && !reverse) continue;
        const load = world.metrics?.roadLoads?.[candidate.ID];
        result.cars += load?.cars ?? 0;
        result.waiting += load?.waiting ?? 0;
    }
    return result;
}

function drawGround(width, height) {
    ctx.fillStyle = "#e9efea";
    ctx.fillRect(0, 0, width, height);

    ctx.strokeStyle = "#dce6de";
    ctx.lineWidth = 1;
    for (let x = 0; x < width; x += 42) {
        ctx.beginPath();
        ctx.moveTo(x, 0);
        ctx.lineTo(x, height);
        ctx.stroke();
    }
    for (let y = 0; y < height; y += 42) {
        ctx.beginPath();
        ctx.moveTo(0, y);
        ctx.lineTo(width, y);
        ctx.stroke();
    }
}

function createView(roads, buildings, width, height) {
    const points = [...Object.values(roads).flatMap(road => road.Points), ...buildings.map(building => building.Position)];
    const minX = Math.min(...points.map(point => point.X));
    const maxX = Math.max(...points.map(point => point.X));
    const minY = Math.min(...points.map(point => point.Y));
    const maxY = Math.max(...points.map(point => point.Y));
    const padding = Math.min(77, Math.max(36, width * 0.075));
    const scale = Math.min(
        (width - padding * 2) / Math.max(maxX - minX, 1),
        (height - padding * 2) / Math.max(maxY - minY, 1)
    );
    return {
        scale,
        x: (width - (maxX - minX) * scale) / 2 - minX * scale,
        y: (height - (maxY - minY) * scale) / 2 - minY * scale
    };
}

function screenPoint(point, view) {
    return { x: view.x + point.X * view.scale, y: view.y + point.Y * view.scale };
}

function roadPath(road, view) {
    if (!road.Points.length) return false;
    ctx.beginPath();
    road.Points.forEach((point, index) => {
        const screen = screenPoint(point, view);
        if (index === 0) ctx.moveTo(screen.x, screen.y);
        else ctx.lineTo(screen.x, screen.y);
    });
    return true;
}

function strokeRoad(road, view, width, color, shadow) {
    if (!roadPath(road, view)) return;
    ctx.save();
    ctx.lineCap = "round";
    ctx.lineJoin = "round";
    ctx.lineWidth = width;
    ctx.strokeStyle = color;
    ctx.shadowColor = "#071519aa";
    ctx.shadowBlur = shadow;
    ctx.stroke();
    ctx.restore();
}

function drawJunction(junction, roads, view, roadWidth) {
    const incoming = roads[junction.IncomingRoads?.[0]];
    const outgoing = roads[junction.OutgoingRoads?.[0]];
    const point = incoming?.Points.at(-1) ?? outgoing?.Points[0];
    if (!point) return;
    const center = screenPoint(point, view);
    ctx.fillStyle = "#9bac9f";
    ctx.fillRect(center.x - roadWidth * 0.56, center.y - roadWidth * 0.56, roadWidth * 1.12, roadWidth * 1.12);
    ctx.fillStyle = "#34464e";
    ctx.fillRect(center.x - roadWidth * 0.5, center.y - roadWidth * 0.5, roadWidth, roadWidth);
}

function drawRoundabout(circle, view, roadWidth) {
    const center = screenPoint(circle.Center, view);
    const radius = circle.Radius * view.scale - roadWidth * 0.57;
    if (radius <= 0) return;
    ctx.save();
    ctx.fillStyle = "#69817e";
    ctx.beginPath();
    ctx.arc(center.x, center.y, radius + 4, 0, Math.PI * 2);
    ctx.fill();
    ctx.fillStyle = "#d2e4d1";
    ctx.beginPath();
    ctx.arc(center.x, center.y, radius, 0, Math.PI * 2);
    ctx.fill();
    ctx.strokeStyle = "#8cae86";
    ctx.lineWidth = 2;
    ctx.beginPath();
    ctx.arc(center.x, center.y, radius * 0.72, 0, Math.PI * 2);
    ctx.stroke();
    ctx.fillStyle = "#537358";
    ctx.font = "700 11px -apple-system, BlinkMacSystemFont, sans-serif";
    ctx.textAlign = "center";
    ctx.textBaseline = "middle";
    ctx.fillText("КРУГ", center.x, center.y);
    ctx.restore();
}

function drawBuilding(building, roads, arrivals, view) {
    const road = roads[building.RoadID];
    if (!road) return;
    const home = building.Kind === "home";
    const center = screenPoint(building.Position, view);
    ctx.save();
    ctx.shadowColor = "#06151ab0";
    ctx.shadowBlur = 8;
    ctx.fillStyle = "#10242b";
    ctx.beginPath();
    ctx.roundRect(center.x - 16, center.y - 20, 32, 39, 8);
    ctx.fill();
    ctx.shadowBlur = 0;
    if (home) {
        ctx.fillStyle = "#57b9ca";
        ctx.beginPath();
        ctx.moveTo(center.x - 12, center.y - 4);
        ctx.lineTo(center.x, center.y - 15);
        ctx.lineTo(center.x + 12, center.y - 4);
        ctx.closePath();
        ctx.fill();
        ctx.fillStyle = "#d8f1ec";
        ctx.fillRect(center.x - 9, center.y - 4, 18, 16);
        ctx.fillStyle = "#5096a4";
        ctx.fillRect(center.x - 2, center.y + 3, 4, 9);
    } else {
        ctx.fillStyle = "#e4aa62";
        ctx.beginPath();
        ctx.roundRect(center.x - 10, center.y - 14, 20, 26, 2);
        ctx.fill();
        ctx.fillStyle = "#5c7890";
        for (let row = 0; row < 3; row++) {
            ctx.fillRect(center.x - 6, center.y - 10 + row * 7, 4, 4);
            ctx.fillRect(center.x + 2, center.y - 10 + row * 7, 4, 4);
        }
    }
    ctx.fillStyle = home ? "#bcebf0" : "#ffdbad";
    ctx.font = "700 9px -apple-system, BlinkMacSystemFont, sans-serif";
    ctx.textAlign = "center";
    ctx.textBaseline = "middle";
    ctx.fillText((home ? "Д" : "Р") + (building.ID % 100), center.x, center.y + 23);
    if (!home && (arrivals[building.ID] ?? 0) > 0) {
        ctx.fillStyle = "#82e2af";
        ctx.beginPath();
        ctx.arc(center.x + 12, center.y - 18, 9, 0, Math.PI * 2);
        ctx.fill();
        ctx.fillStyle = "#10242b";
        ctx.font = "700 9px -apple-system, BlinkMacSystemFont, sans-serif";
        ctx.fillText(String(arrivals[building.ID]), center.x + 12, center.y - 18);
    }
    ctx.restore();
}

function drawTrafficLight(light, roads, view, roadWidth) {
    const road = roads[light.RoadID];
    if (!road) return;
    const position = pointOnRoad(road, light.Position);
    const centerX = view.x + position.x * view.scale;
    const centerY = view.y + position.y * view.scale;
    const normalX = -Math.sin(position.angle);
    const normalY = Math.cos(position.angle);

    ctx.save();
    ctx.strokeStyle = "#e7eee2bd";
    ctx.lineWidth = 3;
    ctx.beginPath();
    ctx.moveTo(centerX, centerY);
    ctx.lineTo(centerX + normalX * roadWidth * 0.45, centerY + normalY * roadWidth * 0.45);
    ctx.stroke();

    const lampX = centerX + normalX * (roadWidth * 0.5 + 15);
    const lampY = centerY + normalY * (roadWidth * 0.5 + 15);
    ctx.strokeStyle = "#76948b";
    ctx.lineWidth = 3;
    ctx.beginPath();
    ctx.moveTo(centerX + normalX * roadWidth * 0.5, centerY + normalY * roadWidth * 0.5);
    ctx.lineTo(lampX, lampY);
    ctx.stroke();
    ctx.fillStyle = "#101d23";
    ctx.strokeStyle = "#739186";
    ctx.lineWidth = 1.5;
    ctx.beginPath();
    ctx.roundRect(lampX - 10, lampY - 22, 20, 44, 7);
    ctx.fill();
    ctx.stroke();

    const bulbs = [
        { y: lampY - 13, color: light.Green ? "#4d302f" : "#ff766b", glow: !light.Green },
        { y: lampY, color: "#615039", glow: false },
        { y: lampY + 13, color: light.Green ? "#75ecaa" : "#294b3d", glow: light.Green }
    ];
    for (const bulb of bulbs) {
        ctx.fillStyle = bulb.color;
        ctx.shadowColor = bulb.color;
        ctx.shadowBlur = bulb.glow ? 14 : 0;
        ctx.beginPath();
        ctx.arc(lampX, bulb.y, 4.1, 0, Math.PI * 2);
        ctx.fill();
    }
    ctx.restore();
}

function drawCar(car, roads, intersections, roundabouts, view) {
    const road = roads[car.Route[car.RoadIndex]];
    if (!road) return;
    const front = lanePointOnRoute(car, car.RoadIndex, car.Position, roads, intersections, roundabouts);
    const rear = carRearPoint(car, roads, intersections, roundabouts);
    const dx = front.x - rear.x;
    const dy = front.y - rear.y;
    const angle = Math.hypot(dx, dy) > 0.001 ? Math.atan2(dy, dx) : front.angle;
    const length = Math.max(8, carLength * view.scale);
    const width = Math.max(5, 2.6 * view.scale);
    const routeColor = car.Route.reduce((sum, id) => sum * 7 + id, 0);
    const color = carColors[Math.abs(routeColor) % carColors.length];

    ctx.save();
    ctx.translate(view.x + front.x * view.scale, view.y + front.y * view.scale);
    ctx.rotate(angle);
    ctx.fillStyle = "#06141877";
    ctx.beginPath();
    ctx.ellipse(-length * 0.48, 3, length * 0.58, width * 0.62, 0, 0, Math.PI * 2);
    ctx.fill();

    ctx.fillStyle = "#17242a";
    ctx.fillRect(-length * 0.82, -width * 0.63, length * 0.19, 3);
    ctx.fillRect(-length * 0.82, width * 0.45, length * 0.19, 3);
    ctx.fillRect(-length * 0.3, -width * 0.63, length * 0.19, 3);
    ctx.fillRect(-length * 0.3, width * 0.45, length * 0.19, 3);

    ctx.fillStyle = color;
    ctx.strokeStyle = "#ffffff70";
    ctx.lineWidth = 1;
    ctx.beginPath();
    ctx.roundRect(-length, -width / 2, length, width, Math.min(5, width * 0.4));
    ctx.fill();
    ctx.stroke();

    ctx.fillStyle = "#203947";
    ctx.beginPath();
    ctx.roundRect(-length * 0.69, -width * 0.36, length * 0.42, width * 0.72, 3);
    ctx.fill();
    ctx.fillStyle = "#9dc6cd8c";
    ctx.fillRect(-length * 0.35, -width * 0.27, Math.max(2, length * 0.05), width * 0.54);
    ctx.fillStyle = "#e9f7db";
    ctx.fillRect(-2.4, -width * 0.32, 2, 3);
    ctx.fillRect(-2.4, width * 0.32 - 3, 2, 3);
    ctx.fillStyle = "#ff6965";
    ctx.fillRect(-length + 0.7, -width * 0.32, 1.8, 3);
    ctx.fillRect(-length + 0.7, width * 0.32 - 3, 1.8, 3);
    ctx.restore();
}

function carRearPoint(car, roads, intersections, roundabouts) {
    let roadIndex = car.RoadIndex;
    let distance = car.Position - carLength;
    while (distance < 0 && roadIndex > 0) {
        roadIndex--;
        distance += roads[car.Route[roadIndex]].Length;
    }
    const road = roads[car.Route[roadIndex]];
    if (distance >= 0) return lanePointOnRoute(car, roadIndex, distance, roads, intersections, roundabouts);
    const start = lanePointOnRoute(car, roadIndex, 0, roads, intersections, roundabouts);
    return {
        x: start.x + Math.cos(start.angle) * distance,
        y: start.y + Math.sin(start.angle) * distance
    };
}

function lanePointOnRoute(car, roadIndex, distance, roads, intersections, roundabouts) {
    const road = roads[car.Route[roadIndex]];
    const position = pointOnRoad(road, distance);
    if (road.Kind === "access") return position;
    let factor = 1;
    if (roads[car.Route[roadIndex - 1]]?.Kind === "access") {
        factor = Math.min(factor, Math.max(0, Math.min(1, distance / 12)));
    }
    if (roads[car.Route[roadIndex + 1]]?.Kind === "access") {
        factor = Math.min(factor, Math.max(0, Math.min(1, (road.Length - distance) / 12)));
    }
    if (road.Level > 0) {
        factor = Math.min(factor, Math.max(0, Math.min(1, distance / 12, (road.Length - distance) / 12)));
    }
    for (const junction of intersections) {
        const movement = junctionMovement(car, roadIndex, junction, roads);
        if (!movement) continue;
        if (movement.right) {
            const turn = rightTurnPointOnRoute(movement, road.ID, distance);
            if (turn) return turn;
        }
        if (movement.straight) continue;
        if (junction.IncomingRoads.includes(road.ID)) {
            factor = Math.min(factor, Math.max(0, Math.min(1, (road.Length - distance) / intersectionZoneLength)));
        }
        if (junction.OutgoingRoads.includes(road.ID)) {
            factor = Math.min(factor, Math.max(0, Math.min(1, distance / intersectionZoneLength)));
        }
    }
    for (const circle of roundabouts) {
        if (circle.Segments.includes(road.ID)) {
            factor = Math.min(factor, Math.max(0, Math.min(1, distance / 10, (road.Length - distance) / 10)));
        }
        for (const entry of circle.Entries) {
            if (entry.ApproachRoadID === road.ID) {
                factor = Math.min(factor, Math.max(0, Math.min(1, (road.Length - distance) / 12)));
            }
            if (entry.ExitRoadID === road.ID) {
                factor = Math.min(factor, Math.max(0, Math.min(1, distance / 12)));
            }
        }
    }
    return {
        x: position.x - Math.sin(position.angle) * laneOffset * factor,
        y: position.y + Math.cos(position.angle) * laneOffset * factor,
        angle: position.angle
    };
}

function junctionMovement(car, roadIndex, junction, roads) {
    const roadID = car.Route[roadIndex];
    const incomingID = junction.IncomingRoads.includes(roadID) ? roadID : car.Route[roadIndex - 1];
    const outgoingID = junction.OutgoingRoads.includes(roadID) ? roadID : car.Route[roadIndex + 1];
    if (!junction.IncomingRoads.includes(incomingID) || !junction.OutgoingRoads.includes(outgoingID)) return null;
    const incoming = roads[incomingID];
    const outgoing = roads[outgoingID];
    if (!incoming || !outgoing || incoming.Points.length < 2 || outgoing.Points.length < 2) return null;
    const arrival = pointOnRoad(incoming, incoming.Length - 0.001);
    const departure = pointOnRoad(outgoing, 0.001);
    const angle = departure.angle - arrival.angle;
    return { incoming, outgoing, arrival, departure, straight: Math.cos(angle) > 0.99, right: Math.sin(angle) > 0.99 };
}

function rightTurnPointOnRoute(movement, roadID, distance) {
    const { incoming, outgoing, arrival, departure } = movement;
    const entering = roadID === incoming.ID;
    if (entering && distance < incoming.Length - intersectionZoneLength) return null;
    if (!entering && distance > intersectionZoneLength) return null;
    const progress = entering
        ? (distance - incoming.Length + intersectionZoneLength) / (intersectionZoneLength * 2)
        : 0.5 + distance / (intersectionZoneLength * 2);
    const t = Math.max(0, Math.min(1, progress));
    const start = pointOnRoad(incoming, incoming.Length - intersectionZoneLength);
    const end = pointOnRoad(outgoing, intersectionZoneLength);
    const center = pointOnRoad(incoming, incoming.Length);
    start.x -= Math.sin(arrival.angle) * laneOffset;
    start.y += Math.cos(arrival.angle) * laneOffset;
    end.x -= Math.sin(departure.angle) * laneOffset;
    end.y += Math.cos(departure.angle) * laneOffset;
    const controlX = center.x - (Math.sin(arrival.angle) + Math.sin(departure.angle)) * laneOffset;
    const controlY = center.y + (Math.cos(arrival.angle) + Math.cos(departure.angle)) * laneOffset;
    const x = (1 - t) ** 2 * start.x + 2 * (1 - t) * t * controlX + t ** 2 * end.x;
    const y = (1 - t) ** 2 * start.y + 2 * (1 - t) * t * controlY + t ** 2 * end.y;
    const tangentX = (1 - t) * (controlX - start.x) + t * (end.x - controlX);
    const tangentY = (1 - t) * (controlY - start.y) + t * (end.y - controlY);
    return { x, y, angle: Math.atan2(tangentY, tangentX) };
}

function pointOnRoad(road, distance) {
    const points = road.Points;
    for (let i = 1; i < points.length; i++) {
        const start = points[i - 1];
        const end = points[i];
        const dx = end.X - start.X;
        const dy = end.Y - start.Y;
        const length = Math.hypot(dx, dy);
        if (!length) continue;
        if (distance <= length) {
            const ratio = distance / length;
            return { x: start.X + dx * ratio, y: start.Y + dy * ratio, angle: Math.atan2(dy, dx) };
        }
        distance -= length;
    }
    const last = points.at(-1);
    const previous = points.at(-2) ?? last;
    return { x: last.X, y: last.Y, angle: Math.atan2(last.Y - previous.Y, last.X - previous.X) };
}

function updateDashboard(world) {
    const custom = Boolean(world.customMap);
    document.getElementById("map-eyebrow").textContent = custom ? "Своя карта" : "Стандартная карта";
    document.getElementById("intro-description").textContent = custom ? "Дороги и здания из редактора. Маршруты рассчитывает Go." : "Машины едут из домов к рабочим местам.";
    document.getElementById("map-badge-right").hidden = custom;
    document.getElementById("stage-footer-message").textContent = "Нажми на дом, работу или дорогу, чтобы посмотреть настройки.";
    carCount.textContent = (world.cars ?? []).length;
    const ringRoads = new Set((world.roundabouts ?? []).flatMap(circle => circle.Segments));
    roadCount.textContent = physicalRoads(Object.values(world.roads ?? {})).filter(road => road.Kind !== "access" && !ringRoads.has(road.ID)).length;
    lightCount.textContent = (world.trafficLights ?? []).filter(light => light.Green).length;
    arrivedCount.textContent = Object.values(world.arrivals ?? {}).reduce((sum, count) => sum + count, 0);
    tripTime.textContent = world.metrics?.completedTrips ? `${Math.round(world.metrics.averageTravelTime)} с` : "—";
    waitTime.textContent = world.metrics?.completedTrips ? `${Math.round(world.metrics.averageWaitTime)} с` : "—";
    if (selectedObject?.kind === "work") {
        const counter = document.getElementById("selected-arrivals");
        if (counter) counter.textContent = String(world.arrivals?.[selectedObject.id] ?? 0);
    }
    if (selectedObject?.kind === "road") {
        const counter = document.getElementById("road-load");
        const road = world.roads[selectedObject.id];
        if (counter && road) {
            const load = physicalRoadLoad(road, world);
            counter.textContent = `${load.cars} / ${load.waiting}`;
        }
    }
    status.textContent = world.paused ? "Симуляция на паузе" : "Симуляция работает";
    pauseButton.disabled = world.paused;
    resumeButton.disabled = !world.paused;
    connection.className = "connection live";
    connection.textContent = "Подключено";
}

async function refresh() {
    try {
        const response = await fetch("/state", { cache: "no-store" });
        if (!response.ok) throw new Error("HTTP " + response.status);
        latestWorld = await response.json();
        updateDashboard(latestWorld);
        draw(latestWorld);
    } catch (error) {
        connection.className = "connection offline";
        connection.textContent = "Нет соединения";
        status.textContent = "Ошибка связи с сервером";
    } finally {
        setTimeout(refresh, 120);
    }
}

async function control(action) {
    const labels = { pause: "Пауза включена", resume: "Движение продолжено", reset: "Симуляция сброшена. Нажми «Продолжить»." };
    try {
        const response = await fetch("/control", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ action })
        });
        if (!response.ok) throw new Error(await response.text());
        feedback.classList.remove("error");
        feedback.textContent = labels[action];
        if (action === "reset") {
            selectedObject = null;
            renderInspector();
        }
    } catch (error) {
        feedback.classList.add("error");
        feedback.textContent = "Не удалось выполнить команду: " + error.message;
    }
}

function distanceToSegment(point, start, end) {
    const dx = end.x - start.x;
    const dy = end.y - start.y;
    const lengthSquared = dx * dx + dy * dy;
    const t = lengthSquared ? Math.max(0, Math.min(1, ((point.x - start.x) * dx + (point.y - start.y) * dy) / lengthSquared)) : 0;
    return Math.hypot(point.x - start.x - dx * t, point.y - start.y - dy * t);
}

function objectAtPosition(x, y) {
    if (!latestWorld?.roads) return null;
    const view = createView(latestWorld.roads, latestWorld.buildings ?? [], canvas.clientWidth, canvas.clientHeight);
    for (const building of latestWorld.buildings ?? []) {
        const point = screenPoint(building.Position, view);
        if (Math.hypot(x - point.x, y - point.y) <= 25) return { kind: building.Kind, id: building.ID };
    }
    let picked = null;
    let closest = Math.max(12, view.scale * 5);
    for (const road of physicalRoads(Object.values(latestWorld.roads)).filter(item => item.Kind !== "access")) {
        for (let i = 1; i < road.Points.length; i++) {
            const distance = distanceToSegment({ x, y }, screenPoint(road.Points[i - 1], view), screenPoint(road.Points[i], view));
            if (distance < closest || (distance === closest && road.Level > (latestWorld.roads[picked?.id]?.Level ?? -1))) {
                closest = distance;
                picked = { kind: "road", id: road.ID };
            }
        }
    }
    return picked;
}

function renderInspector() {
    inspectorFeedback.textContent = "";
    inspectorFeedback.classList.remove("error");
    if (!selectedObject || !latestWorld) {
        inspectorBody.innerHTML = '<h3 class="inspector-title">Ничего не выбрано</h3><p class="inspector-description">Нажми на объект на карте.</p>';
        return;
    }
    if (selectedObject.kind === "work") {
        const building = (latestWorld.buildings ?? []).find(item => item.ID === selectedObject.id);
        if (!building) return;
        inspectorBody.innerHTML = '<h3 class="inspector-title" id="selected-name"></h3><p class="inspector-description">Сюда заканчиваются поездки машин.</p><div class="setting-label"><span>Прибыли на работу</span><span class="setting-value" id="selected-arrivals"></span></div>';
        document.getElementById("selected-name").textContent = building.Name;
        document.getElementById("selected-arrivals").textContent = String(latestWorld.arrivals?.[building.ID] ?? 0);
        return;
    }
    if (selectedObject.kind === "home") {
        const building = (latestWorld.buildings ?? []).find(item => item.ID === selectedObject.id);
        const source = (latestWorld.spawnSources ?? []).find(item => item.homeID === selectedObject.id);
        if (!building || !source) return;
        const rate = source.interval > 0 ? Math.round(60 / source.interval) : 0;
        inspectorBody.innerHTML = '<h3 class="inspector-title" id="selected-name"></h3><p class="inspector-description">Здесь начинаются поездки к рабочим местам.</p><label class="setting-label" for="spawn-rate"><span>Машин в минуту</span><span class="setting-value" id="spawn-rate-value"></span></label><input class="setting-range" id="spawn-rate" type="range" min="0" max="30" step="1"><div class="setting-note">0 — остановить спавн · изменение применяется сразу</div>';
        document.getElementById("selected-name").textContent = building.Name;
        const slider = document.getElementById("spawn-rate");
        const value = document.getElementById("spawn-rate-value");
        slider.value = String(rate);
        value.textContent = String(rate);
        slider.addEventListener("input", () => { value.textContent = slider.value; });
        slider.addEventListener("change", () => saveHomeRate(selectedObject.id, Number(slider.value)));
        return;
    }
    const road = latestWorld.roads[selectedObject.id];
    if (!road) return;
    inspectorBody.innerHTML = '<h3 class="inspector-title" id="selected-name"></h3><p class="inspector-description" id="road-summary"></p><div class="setting-label"><span>На участке / ожидают</span><span class="setting-value" id="road-load"></span></div><label class="setting-label" for="road-speed"><span>Ограничение скорости</span><span class="setting-value" id="road-speed-value"></span></label><input class="setting-range" id="road-speed" type="range" min="5" max="120" step="1"><label class="setting-toggle" for="road-closed"><input id="road-closed" type="checkbox"><span>Закрыть участок</span></label><div class="setting-note" id="road-note"></div>';
    document.getElementById("selected-name").textContent = `Дорога R${road.ID}`;
    const kinds = { highway: "Шоссе", ramp: "Съезд", access: "Подъезд", roundabout: "Круг" };
    document.getElementById("road-summary").textContent = kinds[road.Kind] ?? (road.Level ? "Эстакада" : "Городская дорога");
    const load = physicalRoadLoad(road, latestWorld);
    document.getElementById("road-load").textContent = `${load.cars} / ${load.waiting}`;
    document.getElementById("road-note").textContent = `${road.OneWay ? "Односторонний участок." : "Настройка действует в обоих направлениях."} При закрытии машины ищут открытый объезд.`;
    const slider = document.getElementById("road-speed");
    const value = document.getElementById("road-speed-value");
    slider.value = String(Math.round(road.SpeedLimit * 3.6));
    value.textContent = `${slider.value} км/ч`;
    slider.addEventListener("input", () => { value.textContent = `${slider.value} км/ч`; });
    const roadID = selectedObject.id;
    slider.addEventListener("change", () => saveRoadSettings(roadID, { speedKmh: Number(slider.value) }));
    const closed = document.getElementById("road-closed");
    closed.checked = Boolean(road.Closed);
    closed.addEventListener("change", () => saveRoadSettings(roadID, { closed: closed.checked }));
}

async function saveHomeRate(homeID, carsPerMinute) {
    try {
        const response = await fetch("/settings/home", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ homeID, carsPerMinute })
        });
        if (!response.ok) throw new Error(await response.text());
        inspectorFeedback.classList.remove("error");
        inspectorFeedback.textContent = carsPerMinute ? `Теперь ${carsPerMinute} машин в минуту` : "Спавн этого дома выключен";
    } catch (error) {
        renderInspector();
        inspectorFeedback.classList.add("error");
        inspectorFeedback.textContent = "Не удалось изменить спавн: " + error.message;
    }
}

async function saveRoadSettings(roadID, settings) {
    try {
        const response = await fetch("/settings/road", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ roadID, ...settings })
        });
        if (!response.ok) throw new Error(await response.text());
        inspectorFeedback.classList.remove("error");
        inspectorFeedback.textContent = "Настройка дороги применена";
    } catch (error) {
        renderInspector();
        inspectorFeedback.classList.add("error");
        inspectorFeedback.textContent = "Не удалось изменить дорогу: " + error.message;
    }
}

document.getElementById("pause").addEventListener("click", () => control("pause"));
document.getElementById("resume").addEventListener("click", () => control("resume"));
document.getElementById("reset").addEventListener("click", () => control("reset"));
canvas.addEventListener("pointermove", event => {
    const bounds = canvas.getBoundingClientRect();
    canvas.style.cursor = objectAtPosition(event.clientX - bounds.left, event.clientY - bounds.top) ? "pointer" : "default";
});
canvas.addEventListener("click", event => {
    const bounds = canvas.getBoundingClientRect();
    selectedObject = objectAtPosition(event.clientX - bounds.left, event.clientY - bounds.top);
    renderInspector();
    draw(latestWorld);
});
new ResizeObserver(resizeCanvas).observe(canvas.parentElement);
resizeCanvas();
refresh();
