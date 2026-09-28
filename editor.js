const canvas = document.getElementById("map");
const ctx = canvas.getContext("2d");
const status = document.getElementById("status");
const inspector = document.getElementById("inspector");
const toolHint = document.getElementById("tool-hint");
const width = 800;
const height = 520;
const hints = {
    road: "Нажми начальную и конечную точки дороги. Можно начать или закончить на уже нарисованной дороге.",
    home: "Нажми рядом с точкой дороги, где должен стоять дом.",
    work: "Нажми рядом с точкой дороги, где будет рабочее место.",
    light: "Нажми на точку перекрёстка, чтобы добавить или убрать светофор.",
    roundabout: "Нажми на свободное место. Затем соедини дороги с четырьмя точками круга.",
    select: "Нажми на дорогу или здание, чтобы изменить его параметры.",
    delete: "Нажми на объект, который хочешь удалить. Сохранённая карта не изменится до нажатия «Сохранить»."
};
let map = { nodes: [], streets: [], buildings: [], trafficLights: [], lightSettings: [], roundabouts: [] };
let tool = "road";
let startNode = null;
let pointer = null;
let selected = null;

function view() {
    const w = canvas.clientWidth;
    const h = canvas.clientHeight;
    const scale = Math.min((w - 40) / width, (h - 40) / height);
    return { scale, x: (w - width * scale) / 2, y: (h - height * scale) / 2 };
}

function screen(point) {
    const v = view();
    return { x: v.x + point.x * v.scale, y: v.y + point.y * v.scale };
}

function worldPoint(x, y) {
    const v = view();
    return { x: Math.max(0, Math.min(width, (x - v.x) / v.scale)), y: Math.max(0, Math.min(height, (y - v.y) / v.scale)) };
}

function node(id) { return map.nodes.find(item => item.id === id); }
function nextID(items) { return Math.max(0, ...items.map(item => item.id)) + 1; }

function distanceToSegment(point, start, end) {
    const dx = end.x - start.x;
    const dy = end.y - start.y;
    const lengthSquared = dx * dx + dy * dy;
    const t = lengthSquared ? Math.max(0, Math.min(1, ((point.x - start.x) * dx + (point.y - start.y) * dy) / lengthSquared)) : 0;
    return Math.hypot(point.x - start.x - dx * t, point.y - start.y - dy * t);
}

function nearNode(x, y, radius = 16) {
    let found = null;
    let distance = radius;
    for (const item of map.nodes) {
        const p = screen(item);
        const current = Math.hypot(x - p.x, y - p.y);
        if (current < distance) { found = item; distance = current; }
    }
    return found;
}

function nearestNode(point) {
    let found = null;
    let distance = 80;
    for (const item of map.nodes) {
        const current = Math.hypot(point.x - item.x, point.y - item.y);
        if (current < distance) { found = item; distance = current; }
    }
    return found;
}

function getOrCreateNode(x, y) {
    const existing = nearNode(x, y);
    if (existing) { splitAtNode(existing); return existing; }
    let point = worldPoint(x, y);
    let closest = 12;
    for (const street of map.streets) {
        if (map.roundabouts.some(circle => circle.streetIDs.includes(street.id))) continue;
        const from = node(street.from);
        const to = node(street.to);
        const a = screen(from), b = screen(to);
        const dx = b.x - a.x, dy = b.y - a.y;
        const t = ((x - a.x) * dx + (y - a.y) * dy) / (dx * dx + dy * dy);
        if (t <= 0.03 || t >= 0.97) continue;
        const distance = distanceToSegment({ x, y }, a, b);
        if (distance < closest) {
            closest = distance;
            point = { x: from.x + (to.x - from.x) * t, y: from.y + (to.y - from.y) * t };
        }
    }
    const created = { id: nextID(map.nodes), x: Math.round(point.x), y: Math.round(point.y) };
    map.nodes.push(created);
    splitAtNode(created);
    return created;
}

function splitAtNode(junction) {
    for (const street of [...map.streets]) {
        if (map.roundabouts.some(circle => circle.streetIDs.includes(street.id))) continue;
        const from = node(street.from);
        const to = node(street.to);
        if (from.id === junction.id || to.id === junction.id) continue;
        const dx = to.x - from.x, dy = to.y - from.y;
        const t = ((junction.x - from.x) * dx + (junction.y - from.y) * dy) / (dx * dx + dy * dy);
        if (t <= 0.03 || t >= 0.97 || Math.hypot(junction.x - from.x - dx * t, junction.y - from.y - dy * t) > 4) continue;
        const index = map.streets.findIndex(item => item.id === street.id);
        map.streets.splice(index, 1,
            { ...street, to: junction.id },
            { ...street, id: nextID(map.streets), from: junction.id, to: street.to });
    }
}

function objectAt(x, y) {
    for (const building of map.buildings) {
        const p = screen(building);
        if (Math.hypot(x - p.x, y - p.y) < 18) return { kind: "building", id: building.id };
    }
    for (const circle of map.roundabouts) {
        const center = screen({ x: circle.centerX, y: circle.centerY });
        if (Math.hypot(x - center.x, y - center.y) < Math.max(12, circle.radius * view().scale * 0.65)) return { kind: "roundabout", id: circle.id };
    }
    const found = nearNode(x, y);
    if (found) return { kind: "node", id: found.id };
    for (const street of map.streets) {
        if (map.roundabouts.some(circle => circle.streetIDs.includes(street.id))) continue;
        const from = node(street.from);
        const to = node(street.to);
        if (from && to && distanceToSegment({ x, y }, screen(from), screen(to)) < 12) return { kind: "street", id: street.id };
    }
    return null;
}

function show(message, error = false) {
    status.textContent = message;
    status.classList.toggle("error", error);
}

function changed() {
    show("Изменения не сохранены.");
    renderInspector();
    draw();
}

function draw() {
    const w = canvas.clientWidth;
    const h = canvas.clientHeight;
    ctx.clearRect(0, 0, w, h);
    ctx.fillStyle = "#e9efea";
    ctx.fillRect(0, 0, w, h);
    const v = view();
    ctx.strokeStyle = "#dce6de";
    ctx.lineWidth = 1;
    for (let x = 0; x <= width; x += 40) {
        const sx = v.x + x * v.scale;
        ctx.beginPath(); ctx.moveTo(sx, v.y); ctx.lineTo(sx, v.y + height * v.scale); ctx.stroke();
    }
    for (let y = 0; y <= height; y += 40) {
        const sy = v.y + y * v.scale;
        ctx.beginPath(); ctx.moveTo(v.x, sy); ctx.lineTo(v.x + width * v.scale, sy); ctx.stroke();
    }
    ctx.strokeStyle = "#b9c8bf";
    ctx.strokeRect(v.x, v.y, width * v.scale, height * v.scale);
    for (const circle of map.roundabouts) {
        const center = screen({ x: circle.centerX, y: circle.centerY });
        ctx.lineCap = "round";
        ctx.strokeStyle = selected?.kind === "roundabout" && selected.id === circle.id ? "#267d68" : "#a8b5ac";
        ctx.lineWidth = Math.max(15, v.scale * 13);
        ctx.beginPath(); ctx.arc(center.x, center.y, circle.radius * v.scale, 0, Math.PI * 2); ctx.stroke();
        ctx.strokeStyle = "#3c4b50";
        ctx.lineWidth = Math.max(11, v.scale * 9); ctx.stroke();
        ctx.strokeStyle = "#e6cc8caa";
        ctx.lineWidth = 1.5;
        ctx.setLineDash([9, 12]); ctx.stroke(); ctx.setLineDash([]);
    }
    for (const street of map.streets) {
        if (map.roundabouts.some(circle => circle.streetIDs.includes(street.id))) continue;
        const from = node(street.from);
        const to = node(street.to);
        if (!from || !to) continue;
        const a = screen(from), b = screen(to);
        ctx.lineCap = "round";
        ctx.strokeStyle = selected?.kind === "street" && selected.id === street.id ? "#267d68" : "#a8b5ac";
        ctx.lineWidth = Math.max(15, v.scale * 13);
        ctx.beginPath(); ctx.moveTo(a.x, a.y); ctx.lineTo(b.x, b.y); ctx.stroke();
        ctx.strokeStyle = "#3c4b50";
        ctx.lineWidth = Math.max(11, v.scale * 9);
        ctx.stroke();
        ctx.strokeStyle = "#e6cc8caa";
        ctx.lineWidth = 1.5;
        ctx.setLineDash([9, 12]);
        ctx.stroke();
        ctx.setLineDash([]);
        if (street.oneWay) {
            const angle = Math.atan2(b.y - a.y, b.x - a.x);
            const mid = { x: (a.x + b.x) / 2, y: (a.y + b.y) / 2 };
            ctx.fillStyle = "#f5e7ba";
            ctx.beginPath(); ctx.moveTo(mid.x + Math.cos(angle) * 8, mid.y + Math.sin(angle) * 8);
            ctx.lineTo(mid.x - Math.cos(angle) * 4 - Math.sin(angle) * 5, mid.y - Math.sin(angle) * 4 + Math.cos(angle) * 5);
            ctx.lineTo(mid.x - Math.cos(angle) * 4 + Math.sin(angle) * 5, mid.y - Math.sin(angle) * 4 - Math.cos(angle) * 5); ctx.fill();
        }
    }
    if (startNode && pointer) {
        const a = screen(startNode);
        ctx.strokeStyle = "#267d68";
        ctx.lineWidth = 3;
        ctx.setLineDash([8, 8]);
        ctx.beginPath(); ctx.moveTo(a.x, a.y); ctx.lineTo(pointer.x, pointer.y); ctx.stroke();
        ctx.setLineDash([]);
    }
    for (const item of map.nodes) {
        const p = screen(item);
        ctx.fillStyle = startNode?.id === item.id || selected?.kind === "node" && selected.id === item.id ? "#267d68" : "#557b6d";
        ctx.strokeStyle = "#fff";
        ctx.lineWidth = 2;
        ctx.beginPath(); ctx.arc(p.x, p.y, 6, 0, Math.PI * 2); ctx.fill(); ctx.stroke();
        if (map.trafficLights.includes(item.id)) {
            ctx.fillStyle = "#2b9b70";
            ctx.beginPath(); ctx.arc(p.x + 11, p.y - 11, 6, 0, Math.PI * 2); ctx.fill();
        }
    }
    for (const building of map.buildings) {
        const p = screen(building);
        const junction = node(building.nodeID);
        if (junction) {
            const j = screen(junction);
            ctx.strokeStyle = "#91a399";
            ctx.lineWidth = 4;
            ctx.beginPath(); ctx.moveTo(p.x, p.y); ctx.lineTo(j.x, j.y); ctx.stroke();
        }
        ctx.fillStyle = building.kind === "home" ? "#b8eff1" : "#efba76";
        ctx.strokeStyle = selected?.kind === "building" && selected.id === building.id ? "#267d68" : "#fff";
        ctx.lineWidth = 3;
        ctx.beginPath(); ctx.roundRect(p.x - 11, p.y - 11, 22, 22, 4); ctx.fill(); ctx.stroke();
        ctx.fillStyle = "#263c38";
        ctx.font = "700 14px -apple-system, BlinkMacSystemFont, sans-serif";
        ctx.textAlign = "center";
        ctx.textBaseline = "middle";
        ctx.fillText(building.kind === "home" ? "⌂" : "▣", p.x, p.y);
    }
}

function resize() {
    const ratio = Math.min(window.devicePixelRatio || 1, 2);
    canvas.width = Math.round(canvas.clientWidth * ratio);
    canvas.height = Math.round(canvas.clientHeight * ratio);
    ctx.setTransform(ratio, 0, 0, ratio, 0, 0);
    draw();
}

function renderInspector() {
    inspector.replaceChildren();
    if (!selected) {
        inspector.innerHTML = '<p class="hint">Выбери дорогу или здание, чтобы изменить его параметры.</p>';
        return;
    }
    if (selected.kind === "node") {
        if (!map.trafficLights.includes(selected.id)) {
            inspector.innerHTML = `<h3 class="inspector-name">Точка ${selected.id}</h3><p class="hint">Соедини здесь дороги. Светофор можно поставить отдельным инструментом.</p>`;
            return;
        }
        const setting = map.lightSettings.find(item => item.nodeID === selected.id) ?? { nodeID: selected.id, horizontalGreen: 10, verticalGreen: 10, firstPhase: "horizontal" };
        inspector.innerHTML = `<h3 class="inspector-name">Светофор ${selected.id}</h3><label class="setting" for="horizontal-green">Зелёный по горизонтали, с</label><input id="horizontal-green" class="range" type="number" min="2" max="120"><label class="setting" for="vertical-green">Зелёный по вертикали, с</label><input id="vertical-green" class="range" type="number" min="2" max="120"><label class="setting" for="first-phase">Первый зелёный</label><select id="first-phase" class="range"><option value="horizontal">Горизонтально</option><option value="vertical">Вертикально</option></select><p class="hint">Направления определяются по подходящим дорогам.</p>`;
        for (const [field, id] of [["horizontalGreen", "horizontal-green"], ["verticalGreen", "vertical-green"], ["firstPhase", "first-phase"]]) {
            const input = document.getElementById(id);
            input.value = setting[field];
            input.addEventListener("change", () => {
                const next = { nodeID: selected.id, horizontalGreen: Number(document.getElementById("horizontal-green").value), verticalGreen: Number(document.getElementById("vertical-green").value), firstPhase: document.getElementById("first-phase").value };
                if (next.horizontalGreen < 2 || next.horizontalGreen > 120 || next.verticalGreen < 2 || next.verticalGreen > 120) { show("Длительность фазы должна быть от 2 до 120 секунд.", true); return; }
                map.lightSettings = map.lightSettings.filter(item => item.nodeID !== selected.id);
                map.lightSettings.push(next);
                changed();
            });
        }
        return;
    }
    if (selected.kind === "roundabout") {
        inspector.innerHTML = `<h3 class="inspector-name">Круг ${selected.id}</h3><p class="hint">Четыре въезда. Движение по часовой стрелке, перед въездом нужно уступить машинам на круге. Подведи дороги к точкам на окружности.</p>`;
        return;
    }
    const street = selected.kind === "street" && map.streets.find(item => item.id === selected.id);
    const building = selected.kind === "building" && map.buildings.find(item => item.id === selected.id);
    if (!street && !building) { selected = null; renderInspector(); return; }
    const target = street || building;
    const speed = Boolean(street);
    const editable = speed || building.kind === "home";
    const title = speed ? `Дорога ${street.id}` : `${building.kind === "home" ? "Дом" : "Работа"} ${building.id}`;
    inspector.innerHTML = `<h3 class="inspector-name"></h3>${editable ? '<label class="setting" for="setting-range"><span></span><strong id="setting-value"></strong></label><input id="setting-range" class="range" type="range">' : '<p class="hint">Сюда будут приезжать машины из домов.</p>'}`;
    inspector.querySelector(".inspector-name").textContent = title;
    if (!editable) return;
    const slider = document.getElementById("setting-range");
    const value = document.getElementById("setting-value");
    inspector.querySelector(".setting span").textContent = speed ? "Скорость" : "Машин в минуту";
    slider.min = speed ? "5" : "0";
    slider.max = speed ? "120" : "30";
    slider.value = String(speed ? target.speedKmh : target.carsPerMinute);
    value.textContent = speed ? `${slider.value} км/ч` : slider.value;
    slider.addEventListener("input", () => { value.textContent = speed ? `${slider.value} км/ч` : slider.value; });
    slider.addEventListener("change", () => {
        if (speed) target.speedKmh = Number(slider.value);
        else target.carsPerMinute = Number(slider.value);
        changed();
    });
    if (street) {
        const label = document.createElement("label");
        label.className = "setting";
        label.innerHTML = '<span>Односторонняя</span><input type="checkbox">';
        const input = label.querySelector("input");
        input.checked = Boolean(street.oneWay);
        input.addEventListener("change", () => { street.oneWay = input.checked; changed(); });
        inspector.append(label);
        const note = document.createElement("p");
        note.className = "hint";
        note.textContent = "Направление: от первой точки дороги ко второй.";
        inspector.append(note);
    }
}

function canvasClick(x, y) {
    if (tool === "roundabout") {
        const center = worldPoint(x, y);
        const radius = 30;
        if (center.x < radius + 10 || center.x > width - radius - 10 || center.y < radius + 10 || center.y > height - radius - 10) { show("Поставь круг дальше от края карты.", true); return; }
        if (map.nodes.some(item => Math.hypot(item.x - center.x, item.y - center.y) < radius + 18) || map.roundabouts.some(item => Math.hypot(item.centerX - center.x, item.centerY - center.y) < item.radius + radius + 18)) { show("Здесь слишком близко другие дороги или круг.", true); return; }
        if (map.streets.some(street => {
            const from = node(street.from), to = node(street.to);
            return from && to && distanceToSegment(center, from, to) < radius + 15;
        })) { show("Круг пересекается с существующей дорогой.", true); return; }
        const nodeIDs = [];
        const streetIDs = [];
        let nodeID = nextID(map.nodes);
        let streetID = nextID(map.streets);
        for (let i = 0; i < 4; i++) {
            const angle = -Math.PI / 2 + i * Math.PI / 2;
            map.nodes.push({ id: nodeID, x: center.x + radius * Math.cos(angle), y: center.y + radius * Math.sin(angle) });
            nodeIDs.push(nodeID++);
        }
        for (let i = 0; i < 4; i++) {
            map.streets.push({ id: streetID, from: nodeIDs[i], to: nodeIDs[(i + 1) % 4], speedKmh: 25, oneWay: true });
            streetIDs.push(streetID++);
        }
        const circle = { id: nextID(map.roundabouts), centerX: center.x, centerY: center.y, radius, nodeIDs, streetIDs };
        map.roundabouts.push(circle);
        selected = { kind: "roundabout", id: circle.id };
        changed();
        return;
    }
    if (tool === "road") {
        const endpoint = getOrCreateNode(x, y);
        if (!startNode) {
            startNode = endpoint;
            pointer = { x, y };
            show("Теперь нажми конечную точку дороги.");
            draw();
            return;
        }
        if (endpoint.id === startNode.id || Math.hypot(endpoint.x - startNode.x, endpoint.y - startNode.y) < 10) {
            show("Дорога должна соединять две разные точки на расстоянии хотя бы 10 м.", true);
            return;
        }
        if (map.streets.some(item => item.from === startNode.id && item.to === endpoint.id || item.to === startNode.id && item.from === endpoint.id)) {
            show("Эти точки уже соединены дорогой.", true);
            return;
        }
        const street = { id: nextID(map.streets), from: startNode.id, to: endpoint.id, speedKmh: 36 };
        map.streets.push(street);
        selected = { kind: "street", id: street.id };
        startNode = null;
        pointer = null;
        changed();
        return;
    }
    if (tool === "home" || tool === "work") {
        const position = worldPoint(x, y);
        const junction = nearestNode(position);
        if (!junction) { show("Поставь здание рядом с точкой дороги.", true); return; }
        if (Math.hypot(position.x - junction.x, position.y - junction.y) < 5) { show("Поставь здание немного в стороне от точки дороги.", true); return; }
        const building = { id: nextID(map.buildings), kind: tool, nodeID: junction.id, x: Math.round(position.x), y: Math.round(position.y), carsPerMinute: tool === "home" ? 5 : 0 };
        map.buildings.push(building);
        selected = { kind: "building", id: building.id };
        changed();
        return;
    }
    if (tool === "light") {
        const junction = nearNode(x, y);
        if (!junction) { show("Нажми на точку, где соединяются дороги.", true); return; }
        const index = map.trafficLights.indexOf(junction.id);
        if (index < 0) map.trafficLights.push(junction.id);
        else { map.trafficLights.splice(index, 1); map.lightSettings = map.lightSettings.filter(item => item.nodeID !== junction.id); }
        selected = { kind: "node", id: junction.id };
        changed();
        return;
    }
    const found = objectAt(x, y);
    if (tool === "select") {
        selected = found;
        renderInspector();
        draw();
        return;
    }
    if (!found) return;
    if (found.kind === "roundabout") {
        const circle = map.roundabouts.find(item => item.id === found.id);
        map.streets = map.streets.filter(item => !circle.streetIDs.includes(item.id));
        map.roundabouts = map.roundabouts.filter(item => item.id !== circle.id);
        for (const id of circle.nodeIDs) {
            if (!map.streets.some(item => item.from === id || item.to === id) && !map.buildings.some(item => item.nodeID === id)) {
                map.nodes = map.nodes.filter(item => item.id !== id);
                map.trafficLights = map.trafficLights.filter(item => item !== id);
                map.lightSettings = map.lightSettings.filter(item => item.nodeID !== id);
            }
        }
    }
    if (found.kind === "building") map.buildings = map.buildings.filter(item => item.id !== found.id);
    if (found.kind === "street") map.streets = map.streets.filter(item => item.id !== found.id);
    if (found.kind === "node") {
        if (map.streets.some(item => item.from === found.id || item.to === found.id) || map.buildings.some(item => item.nodeID === found.id)) {
            show("Сначала удали дороги и здания, соединённые с этой точкой.", true);
            return;
        }
        map.nodes = map.nodes.filter(item => item.id !== found.id);
        map.trafficLights = map.trafficLights.filter(id => id !== found.id);
        map.lightSettings = map.lightSettings.filter(item => item.nodeID !== found.id);
    }
    selected = null;
    changed();
}

async function saveMap() {
    const response = await fetch("/map", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify(map) });
    if (!response.ok) throw new Error((await response.text()).trim());
    show("Карта сохранена.");
}

canvas.addEventListener("pointerdown", event => {
    const bounds = canvas.getBoundingClientRect();
    canvasClick(event.clientX - bounds.left, event.clientY - bounds.top);
});
canvas.addEventListener("pointermove", event => {
    if (!startNode) return;
    const bounds = canvas.getBoundingClientRect();
    pointer = { x: event.clientX - bounds.left, y: event.clientY - bounds.top };
    draw();
});
for (const button of document.querySelectorAll("[data-tool]")) {
    button.addEventListener("click", () => {
        tool = button.dataset.tool;
        startNode = null;
        pointer = null;
        document.querySelectorAll("[data-tool]").forEach(item => item.classList.toggle("active", item === button));
        toolHint.textContent = hints[tool];
        draw();
    });
}
document.getElementById("save").addEventListener("click", async () => {
    try { await saveMap(); } catch (error) { show("Не удалось сохранить карту: " + error.message, true); }
});
document.getElementById("run").addEventListener("click", async () => {
    try {
        await saveMap();
        const response = await fetch("/map/run", { method: "POST" });
        if (!response.ok) throw new Error((await response.text()).trim());
        location.href = "/";
    } catch (error) { show("Не удалось запустить карту: " + error.message, true); }
});
document.getElementById("default").addEventListener("click", async () => {
    try {
        const response = await fetch("/map/default", { method: "POST" });
        if (!response.ok) throw new Error((await response.text()).trim());
        location.href = "/";
    } catch (error) { show("Не удалось открыть стандартную карту: " + error.message, true); }
});
document.getElementById("clear").addEventListener("click", () => {
    map = { nodes: [], streets: [], buildings: [], trafficLights: [], lightSettings: [], roundabouts: [] };
    selected = null;
    startNode = null;
    pointer = null;
    changed();
});
new ResizeObserver(resize).observe(canvas);
fetch("/map", { cache: "no-store" }).then(async response => {
    if (!response.ok) throw new Error((await response.text()).trim());
    map = await response.json();
    map.nodes ??= [];
    map.streets ??= [];
    map.buildings ??= [];
    map.trafficLights ??= [];
    map.lightSettings ??= [];
    map.roundabouts ??= [];
    draw();
    if (map.streets.length) show("Сохранённая карта загружена. Можно продолжить редактирование.");
}).catch(error => show("Не удалось загрузить карту: " + error.message, true));
