let conflicts = [];
let currentConflict = null;

// Mock data integration check
// (In production, this fetches from Go server)
async function fetchConflicts() {
    try {
        const res = await fetch('/api/conflicts');
        if (!res.ok) throw new Error('API Failed');
        const data = await res.json();
        conflicts = data;
        renderList();
    } catch (e) {
        console.warn("Using mock data due to API error:", e);
        // Fallback for dev mode without backend
        conflicts = [
            {
                alias: "gc",
                existing: { package: "git-essentials", command: "git commit -v" },
                new: { package: "google-cloud", command: "gcloud config" }
            },
            {
                alias: "run",
                existing: { package: "node-scripts", command: "npm run dev" },
                new: { package: "deno-tools", command: "deno run -A" }
            }
        ];
        renderList();
    }
}

function renderList() {
    const list = document.getElementById('conflict-list');
    list.innerHTML = '';

    if (conflicts.length === 0) {
        list.innerHTML = '<li style="padding: 20px; text-align: center; color: var(--success);">All Clear! 🎉</li>';
        document.getElementById('detail-view').style.display = 'none';
        document.getElementById('empty-state').innerHTML = '<h2>All Conflicts Resolved</h2><button class="btn btn-primary" id="shutdown-btn" style="margin-top: 8px;" onclick="shutdown()">Close & Exit</button>';
        document.getElementById('empty-state').style.display = 'block';
        return;
    }

    conflicts.forEach((c, idx) => {
        const li = document.createElement('li');
        li.className = 'conflict-item glass';
        li.innerHTML = `
            <span class="conflict-alias">${c.alias}</span>
            <div class="conflict-pkgs">${c.existing.package} vs ${c.new.package}</div>
        `;
        li.onclick = () => selectConflict(idx);
        list.appendChild(li);
    });
}

function selectConflict(idx) {
    currentConflict = conflicts[idx];

    // Highlight sidebar
    const items = document.querySelectorAll('.conflict-item');
    items.forEach(i => i.classList.remove('active'));
    items[idx].classList.add('active');

    // Update Detail View
    document.getElementById('empty-state').style.display = 'none';
    const detail = document.getElementById('detail-view');
    detail.style.display = 'flex';

    // Populate Data
    // Use textContent/innerText for user-provided strings to prevent XSS
    const title = document.getElementById('alias-title');
    title.innerHTML = 'alias <span style="color: var(--accent-secondary)"></span>';
    title.querySelector('span').innerText = currentConflict.alias;

    document.getElementById('pkg-existing-name').innerText = currentConflict.existing.package;
    document.getElementById('cmd-existing').innerText = currentConflict.existing.command;

    document.getElementById('pkg-new-name').innerText = currentConflict.new.package;
    document.getElementById('cmd-new').innerText = currentConflict.new.command;
}

async function resolve(action) {
    if (!currentConflict) return;

    // Optimistic UI update
    const note = document.createElement('div');
    note.className = 'glass';
    note.style = 'position: fixed; bottom: 20px; right: 20px; padding: 16px; border-radius: 8px; background: var(--success); color: white; animation: slideIn 0.3s;';
    note.innerText = 'Resolved: ' + action;
    document.body.appendChild(note);
    setTimeout(() => note.remove(), 2000);

    // Call API
    try {
        await fetch('/api/resolve', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                alias: currentConflict.alias,
                action: action,
                targetPackage: currentConflict.new.package
            })
        });
    } catch (e) { console.error(e); }

    // Remove from list
    conflicts = conflicts.filter(c => c.alias !== currentConflict.alias);
    renderList();
    if (conflicts.length > 0) selectConflict(0);
}

function renameAlias() {
    const newName = prompt("Enter new name for the incoming alias:", currentConflict.alias + "_new");
    if (newName && newName !== currentConflict.alias) {
        // Send rename action
        resolve('rename:' + newName);
    }
}

async function shutdown() {
    const btn = document.getElementById('shutdown-btn');
    btn.innerText = 'Closing...';
    btn.disabled = true;

    try {
        await fetch('/api/shutdown');
        document.body.innerHTML = `
            <div style="height: 100vh; display: flex; flex-direction: column; align-items: center; justify-content: center; color: white; background: #0d0d12; font-family: sans-serif;">
                <h1 style="color: #22c55e;">Done!</h1>
                <p style="color: #a1a1aa;">You can now close this tab.</p>
            </div>
        `;
        window.close();
    } catch (e) {
        console.error(e);
    }
}

// Init
fetchConflicts();
