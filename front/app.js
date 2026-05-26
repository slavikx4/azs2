"use strict";

let currentSection = "vehicles";
let currentSort = { key: "id", asc: true };
let searchQuery = "";
let currentReportMonth = "2026-05";
let vehiclesData = [];

window.navigateTo = navigateTo;
window.toggleSidebar = toggleSidebar;
window.handleSearch = handleSearch;
window.openHistory = openHistory;
window.closeModal = closeModal;
window.loadReport = loadReport;
window.openAddVehicleModal = openAddVehicleModal;
window.closeAddVehicleModal = closeAddVehicleModal;
window.handleAddVehicle = handleAddVehicle;
window.updateLimit = updateLimit;
window.updateCardBalance = updateCardBalance;

function navigateTo(section) {
    currentSection = section;
    document.querySelectorAll(".nav-btn").forEach(btn => btn.classList.toggle("active", btn.dataset.section === section));
    document.querySelectorAll(".section").forEach(sec => sec.classList.toggle("active", sec.id === `section-${section}`));
    document.getElementById("page-title").textContent = section === "vehicles" ? "Список машин" : "Отчётность";
    document.getElementById("sidebar").classList.remove("open");
    console.log(`Navigated to: ${section}`);
}

function toggleSidebar() {
    document.getElementById("sidebar").classList.toggle("open");
    console.log("Sidebar toggled");
}

async function fetchVehicles() {
    try {
        console.log("Fetching vehicles...");
        const res = await fetch('/api/vehicles');
        if (!res.ok) {
            throw new Error(`HTTP ${res.status}: ${res.statusText}`);
        }
        const vehicles = await res.json();
        console.log(`Received ${vehicles.length} vehicles`);

        vehiclesData = [];
        for (const v of vehicles) {
            try {
                const cardsRes = await fetch(`/api/vehicles/${v.id}/cards`);
                if (!cardsRes.ok) {
                    throw new Error(`HTTP ${cardsRes.status}: ${cardsRes.statusText}`);
                }
                const cards = await cardsRes.json();
                let balance = v.route_limit;
                if (cards && cards.length > 0) {
                    const lukoilCard = cards.find(c => c.provider_id === 'lukoil');
                    if (lukoilCard) {
                        balance = lukoilCard.balance;
                        console.log(`Vehicle ${v.id} balance: ${balance} (from lukoil card)`);
                    }
                }
                vehiclesData.push({
                    ...v,
                    card_balance: balance,
                    card_id: v.system_card_id
                });
            } catch (err) {
                console.error(`Failed to fetch cards for vehicle ${v.id}:`, err);
                vehiclesData.push({
                    ...v,
                    card_balance: v.route_limit,
                    card_id: v.system_card_id
                });
            }
        }
        renderVehicles();
    } catch (err) {
        console.error("Failed to fetch vehicles:", err);
        alert("Ошибка загрузки списка машин: " + err.message);
    }
}

function renderVehicles() {
    console.log("Rendering vehicles...");
    const tbody = document.getElementById("vehicles-tbody");
    let vehicles = [...vehiclesData];

    if (searchQuery.trim()) {
        const q = searchQuery.toLowerCase();
        vehicles = vehicles.filter(v =>
            v.plate.toLowerCase().includes(q) ||
            String(v.id).includes(q) ||
            (v.card_id && v.card_id.toLowerCase().includes(q))
        );
        console.log(`Filtered to ${vehicles.length} vehicles (search: "${searchQuery}")`);
    }

    vehicles.sort((a, b) => {
        let aVal = a[currentSort.key];
        let bVal = b[currentSort.key];
        if (currentSort.key === "balance") {
            aVal = a.card_balance;
            bVal = b.card_balance;
        }
        const cmp = typeof aVal === "number" && typeof bVal === "number" ? aVal - bVal : String(aVal).localeCompare(String(bVal));
        return currentSort.asc ? cmp : -cmp;
    });

    tbody.innerHTML = "";
    for (const v of vehicles) {
        const fuel = v.fuel_level || 50;
        const balance = v.card_balance !== undefined ? v.card_balance : v.route_limit;
        const cls = fuel > 50 ? "high" : fuel > 20 ? "mid" : "low";
        const badge = fuel > 50 ? "badge-green" : fuel > 20 ? "badge-yellow" : "badge-red";
        const label = fuel < 20 ? "Крит." : fuel < 50 ? "Среднее" : "Норма";

        const row = document.createElement("tr");
        row.onclick = () => openHistory(v.id);

        row.innerHTML = `
            <td><strong>#${v.id}</strong></td>
            <td><span class="live-dot"></span>${v.plate}</td>
            <td>
                <div class="fuel-bar-wrap">
                    <div class="fuel-bar">
                        <div class="fuel-bar-fill ${cls}" style="width:${fuel}%"></div>
                    </div>
                    <span class="fuel-text">${fuel}%</span>
                    <span class="badge ${badge}">${label}</span>
                </div>
            </td>
            <td>
                <strong>${v.route_limit}</strong> л
                <button class="action-btn" onclick="event.stopPropagation(); updateLimit(${v.id}, ${v.route_limit})" title="Изменить лимит">✏️</button>
            </td>
            <td>
                <strong style="color:${balance < 30 ? 'var(--danger)' : 'var(--success)'}">${balance}</strong> л
                <button class="action-btn" onclick="event.stopPropagation(); updateCardBalance(${v.id}, ${balance})" title="Изменить баланс карты">💳</button>
            </td>
            <td><code style="background:#f1f5f9;padding:2px 8px;border-radius:6px;font-size:13px">${v.card_id || '—'}</code></td>
        `;
        tbody.appendChild(row);
    }
    console.log(`Rendered ${tbody.children.length} vehicles`);
}

function setupSorting() {
    document.querySelectorAll("#vehicles-table th").forEach(th => {
        th.addEventListener("click", () => {
            const key = th.dataset.sort;
            if (key) {
                currentSort = currentSort.key === key ? { key, asc: !currentSort.asc } : { key, asc: true };
                console.log(`Sorting by ${key}, ascending: ${currentSort.asc}`);
                renderVehicles();
            }
        });
    });
}

function handleSearch(value) {
    searchQuery = value;
    console.log(`Search query: "${searchQuery}"`);
    renderVehicles();
}

async function openHistory(vehicleId) {
    console.log(`Opening history for vehicle ${vehicleId}`);
    const vehicle = vehiclesData.find(v => v.id === vehicleId);
    if (!vehicle) {
        console.error(`Vehicle ${vehicleId} not found`);
        return;
    }

    try {
        const res = await fetch(`/api/vehicles/${vehicleId}/refuelings`);
        if (!res.ok) {
            throw new Error(`HTTP ${res.status}: ${res.statusText}`);
        }
        const history = await res.json();
        console.log(`Received ${history.length} refueling records`);

        document.getElementById("modal-title").textContent = `История заправок — ${vehicle.plate}`;
        const tbody = document.getElementById("history-tbody");
        let totalLiters = 0, totalSum = 0;

        if (history.length === 0) {
            tbody.innerHTML = '<tr><td colspan="5" style="text-align:center">Нет заправок</td></tr>';
        } else {
            tbody.innerHTML = "";
            for (const r of history) {
                const sum = r.liters * r.price_per_liter;
                totalLiters += r.liters;
                totalSum += sum;
                const row = document.createElement("tr");
                row.innerHTML = `
                    <td>${new Date(r.timestamp).toLocaleDateString("ru-RU")}</td>
                    <td><strong>${r.liters}</strong> л</td>
                    <td>${r.price_per_liter.toFixed(2)} ₽</td>
                    <td><strong>${sum.toFixed(2)}</strong> ₽</td>
                    <td>${r.station_name || ''}<br><small style="color:#94a3b8">${r.address || ''}</small></td>
                `;
                tbody.appendChild(row);
            }
        }

        document.getElementById("modal-total").textContent = `Итого: ${totalLiters.toFixed(1)} л  |  ${totalSum.toFixed(2)} ₽`;
        document.getElementById("modal-overlay").classList.add("open");
        console.log(`History modal opened, total: ${totalLiters}L, ${totalSum.toFixed(2)}₽`);
    } catch (err) {
        console.error("Failed to load history:", err);
        alert("Ошибка загрузки истории: " + err.message);
    }
}

function closeModal(event) {
    if (event && event.target !== event.currentTarget) return;
    document.getElementById("modal-overlay").classList.remove("open");
    console.log("Modal closed");
}

function openAddVehicleModal() {
    console.log("Opening add vehicle modal");
    document.getElementById("modal-add-vehicle").classList.add("open");
}

function closeAddVehicleModal(event) {
    if (event && event.target !== event.currentTarget) return;
    document.getElementById("modal-add-vehicle").classList.remove("open");
    document.getElementById("add-vehicle-form").reset();
    console.log("Add vehicle modal closed");
}

async function handleAddVehicle(event) {
    event.preventDefault();
    console.log("Handling add vehicle form submission");

    const plate = document.getElementById("new-plate").value.trim();
    const limit = parseFloat(document.getElementById("new-limit").value);

    if (!plate || isNaN(limit) || limit <= 0) {
        console.warn("Invalid form data:", { plate, limit });
        alert("Заполните все поля корректно");
        return;
    }

    console.log("Adding vehicle:", { plate, limit });

    try {
        const res = await fetch('/api/vehicles/add', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ plate: plate, limit: limit })
        });

        const data = await res.json();

        if (res.ok) {
            console.log("Vehicle added successfully:", data);
            alert("Машина успешно добавлена");
            closeAddVehicleModal();
            await fetchVehicles();
            await loadReport(currentReportMonth);
        } else {
            console.error("Failed to add vehicle:", data);
            alert("Ошибка добавления: " + (data.message || data.error || "Unknown error"));
        }
    } catch (err) {
        console.error("Network error while adding vehicle:", err);
        alert("Ошибка сети: " + err.message);
    }
}

async function updateLimit(vehicleId, currentLimit) {
    console.log(`Updating limit for vehicle ${vehicleId}, current: ${currentLimit}`);
    const newLimit = prompt("Введите новый лимит (л):", currentLimit);
    if (newLimit === null) {
        console.log("Limit update cancelled");
        return;
    }
    const limit = parseFloat(newLimit);
    if (isNaN(limit) || limit <= 0) {
        console.warn("Invalid limit value:", newLimit);
        alert("Введите корректное число");
        return;
    }

    console.log(`Setting new limit: ${limit}`);

    try {
        const res = await fetch(`/api/vehicles/${vehicleId}/limit`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ limit: limit })
        });
        const data = await res.json();

        if (res.ok) {
            console.log("Limit updated successfully:", data);
            alert("Лимит обновлён");
            await fetchVehicles();
            await loadReport(currentReportMonth);
        } else {
            console.error("Failed to update limit:", data);
            alert("Ошибка обновления: " + (data.message || data.error || "Unknown error"));
        }
    } catch (err) {
        console.error("Network error while updating limit:", err);
        alert("Ошибка сети: " + err.message);
    }
}

async function updateCardBalance(vehicleId, currentBalance) {
    console.log(`Updating balance for vehicle ${vehicleId}, current: ${currentBalance}`);
    const newBalance = prompt("Введите новый баланс карты (л):", currentBalance);
    if (newBalance === null) {
        console.log("Balance update cancelled");
        return;
    }
    const balance = parseFloat(newBalance);
    if (isNaN(balance) || balance < 0) {
        console.warn("Invalid balance value:", newBalance);
        alert("Введите корректное число");
        return;
    }

    console.log(`Setting new balance: ${balance}`);

    try {
        const res = await fetch(`/api/vehicles/${vehicleId}/balance`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ balance: balance })
        });
        const data = await res.json();

        if (res.ok) {
            console.log("Balance updated successfully:", data);
            alert("Баланс обновлён");
            const vehicle = vehiclesData.find(v => v.id === vehicleId);
            if (vehicle) {
                vehicle.card_balance = balance;
            }
            renderVehicles();
            await loadReport(currentReportMonth);
        } else {
            console.error("Failed to update balance:", data);
            alert("Ошибка обновления: " + (data.message || data.error || "Unknown error"));
        }
    } catch (err) {
        console.error("Network error while updating balance:", err);
        alert("Ошибка сети: " + err.message);
    }
}

async function loadReport(month) {
    if (!month) {
        month = currentReportMonth;
    }
    currentReportMonth = month;
    const [year, mon] = month.split("-");
    console.log(`Loading report for ${year}-${mon}`);

    try {
        const res = await fetch(`/api/reports/monthly?year=${year}&month=${mon}`);
        if (!res.ok) {
            throw new Error(`HTTP ${res.status}: ${res.statusText}`);
        }
        let records = await res.json();

        if (!records) {
            records = [];
        }
        console.log(`Received ${records.length} report records`);

        const totalLiters = records.reduce((s, r) => s + (r.liters || 0), 0);
        const totalSpent = records.reduce((s, r) => s + ((r.liters || 0) * (r.price_per_liter || 0)), 0);
        const avgPrice = totalLiters > 0 ? totalSpent / totalLiters : 0;

        const litersEl = document.getElementById("report-liters");
        const avgPriceEl = document.getElementById("report-avg-price");
        const totalSpentEl = document.getElementById("report-total-spent");
        const refuelCountEl = document.getElementById("report-refuel-count");

        if (litersEl) litersEl.textContent = `${totalLiters.toFixed(1)} л`;
        if (avgPriceEl) avgPriceEl.textContent = `${avgPrice.toFixed(2)} ₽`;
        if (totalSpentEl) totalSpentEl.textContent = `${totalSpent.toFixed(2)} ₽`;
        if (refuelCountEl) refuelCountEl.textContent = String(records.length);

        console.log(`Report totals: ${totalLiters.toFixed(1)}L, ${totalSpent.toFixed(2)}₽, avg: ${avgPrice.toFixed(2)}₽, count: ${records.length}`);

        const map = new Map();
        if (vehiclesData && vehiclesData.length > 0) {
            for (const v of vehiclesData) {
                map.set(v.plate, { liters: 0, spent: 0, count: 0 });
            }
        }

        if (records && records.length > 0) {
            for (const r of records) {
                const vehicle = vehiclesData ? vehiclesData.find(v => v.id === r.vehicle_id) : null;
                if (vehicle && map.has(vehicle.plate)) {
                    const ex = map.get(vehicle.plate);
                    ex.liters += r.liters || 0;
                    ex.spent += (r.liters || 0) * (r.price_per_liter || 0);
                    ex.count++;
                    map.set(vehicle.plate, ex);
                }
            }
        }

        const tbody = document.getElementById("report-tbody");
        if (tbody) {
            tbody.innerHTML = "";
            const nonEmpty = Array.from(map.entries()).filter(([_, d]) => d.count > 0);

            if (nonEmpty.length === 0) {
                const row = document.createElement("tr");
                row.innerHTML = '<td colspan="5" style="text-align:center">Нет данных за выбранный месяц</td>';
                tbody.appendChild(row);
                console.log("No data for selected month");
            } else {
                for (const [plate, d] of nonEmpty) {
                    const row = document.createElement("tr");
                    row.innerHTML = `
                        <td><strong>${plate}</strong></td>
                        <td>${d.liters.toFixed(1)} л</td>
                        <td>${d.spent.toFixed(2)} ₽</td>
                        <td>${d.liters > 0 ? (d.spent / d.liters).toFixed(2) : "0.00"} ₽/л</td>
                        <td>${d.count}</td>
                    `;
                    tbody.appendChild(row);
                }
                console.log(`Rendered ${nonEmpty.length} vehicle rows in report`);
            }
        }
    } catch (err) {
        console.error("Failed to load report:", err);
        const tbody = document.getElementById("report-tbody");
        if (tbody) {
            tbody.innerHTML = '<tr><td colspan="5" style="text-align:center">Ошибка загрузки данных</td></tr>';
        }
        alert("Ошибка загрузки отчёта: " + err.message);
    }
}

async function init() {
    console.log("Initializing FuelControl app...");
    document.querySelector(".current-date").textContent = new Date().toLocaleDateString("ru-RU", {
        weekday: "long",
        day: "numeric",
        month: "long",
        year: "numeric"
    });
    document.getElementById("report-month").value = currentReportMonth;

    await fetchVehicles();
    setupSorting();
    await loadReport(currentReportMonth);

    document.getElementById("report-month").addEventListener("change", (e) => {
        console.log("Month changed to:", e.target.value);
        loadReport(e.target.value);
    });

    console.log("FuelControl app initialized successfully");
}

document.addEventListener("DOMContentLoaded", init);