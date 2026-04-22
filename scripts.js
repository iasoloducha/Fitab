// Header scroll effect
const header = document.getElementById('header');
window.addEventListener('scroll', () => {
    if (window.scrollY > 100) {
        header.classList.add('scrolled');
    } else {
        header.classList.remove('scrolled');
    }
});

// Mobile menu toggle
const navToggle = document.getElementById('navToggle');
const navMenu = document.getElementById('navMenu');

navToggle.addEventListener('click', () => {
    navToggle.classList.toggle('active');
    navMenu.classList.toggle('active');
});

// Close menu on link click
document.querySelectorAll('.nav-menu a').forEach(link => {
    link.addEventListener('click', () => {
        navToggle.classList.remove('active');
        navMenu.classList.remove('active');
    });
});

// Smooth scroll for anchor links
document.querySelectorAll('a[href^="#"]').forEach(anchor => {
    anchor.addEventListener('click', function(e) {
        e.preventDefault();
        const target = document.querySelector(this.getAttribute('href'));
        if (target) {
            target.scrollIntoView({
                behavior: 'smooth',
                block: 'start'
            });
        }
    });
});

// Add exercise functionality
function addExercise(containerId) {
    const container = document.getElementById(containerId);
    const newRow = document.createElement('div');
    newRow.className = 'exercise-row';

    // Build inputs with DOM methods to avoid XSS
    const nameInput = document.createElement('input');
    nameInput.type = 'text';
    nameInput.className = 'exercise-name';
    nameInput.placeholder = 'Nombre del ejercicio';
    newRow.appendChild(nameInput);

    const setsInput = document.createElement('input');
    setsInput.type = 'text';
    setsInput.className = 'exercise-sets';
    setsInput.placeholder = 'series';
    newRow.appendChild(setsInput);

    const repsInput = document.createElement('input');
    repsInput.type = 'text';
    repsInput.className = 'exercise-reps';
    repsInput.placeholder = 'rep';
    newRow.appendChild(repsInput);

    const restInput = document.createElement('input');
    restInput.type = 'text';
    restInput.className = 'exercise-obs';
    restInput.placeholder = 'obs';
    newRow.appendChild(restInput);

    const removeBtn = document.createElement('button');
    removeBtn.type = 'button';
    removeBtn.className = 'remove-exercise';
    removeBtn.setAttribute('aria-label', 'Eliminar ejercicio');
    removeBtn.textContent = '🗑';
    newRow.appendChild(removeBtn);

    // Add row to container
    container.appendChild(newRow);

    // Add remove event listener
    removeBtn.addEventListener('click', function() {
        newRow.remove();
    });
}

// Remove exercise functionality
document.querySelectorAll('.remove-exercise').forEach(btn => {
    btn.addEventListener('click', function() {
        this.parentElement.remove();
    });
});

// Save routine - Open print dialog
function saveRoutine() {
    const nombre = document.getElementById('alumno-nombre').value || 'Alumno';
    const edad = document.getElementById('alumno-edad').value;
    const objetivo = document.getElementById('alumno-objetivo').value;
    const periodo = document.getElementById('alumno-periodo').value;
    const periodoFormateado = periodo ? new Date(periodo).toLocaleDateString('es-AR') : '';

    const objetivoLabels = { 'bajar-peso': 'Bajar peso', 'ganar-musculo': 'Ganar músculo', 'mantenerse': 'Mantenerse', 'resistencia': 'Resistencia' };

    // Build HTML for print
    let html = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Rutina - ${nombre}</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: 'Helvetica Neue', Arial, sans-serif; font-size: 11pt; color: #000; }
        .header { border: 2px solid #000; padding: 20px; margin-bottom: 20px; }
        .header h1 { font-size: 24pt; font-weight: bold; }
        .header p { font-size: 11pt; color: #333; margin-top: 5px; }
        .info-box { border: 1px solid #000; padding: 15px 20px; margin-bottom: 25px; }
        .info-label { font-weight: bold; }
        .day-section { margin-bottom: 25px; page-break-inside: avoid; }
        .day-header { border: 2px solid #000; border-bottom: none; padding: 10px 15px; font-size: 12pt; font-weight: bold; }
        table { width: 100%; border-collapse: collapse; border: 2px solid #000; }
        th { border: 1px solid #000; padding: 8px; font-size: 9pt; font-weight: bold; text-align: left; }
        td { border: 1px solid #000; padding: 8px; }
        .col-num { width: 30px; text-align: center; }
        .col-sets { width: 60px; text-align: center; }
        .col-reps { width: 80px; text-align: center; }
        .col-rest { width: 80px; text-align: center; }
        @media print {
            body { -webkit-print-color-adjust: exact; print-color-adjust: exact; }
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>FITLAB</h1>
        <p>Rutina de Entrenamiento</p>
    </div>

    <div class="info-box">
        <p><span class="info-label">Alumno:</span> ${nombre.toUpperCase()}</span> ${edad ? `<span class="info-label">Edad:</span> ${edad} años` : ''} ${objetivo ? `<span class="info-label">Objetivo:</span> ${objetivoLabels[objetivo] || objetivo}` : ''} ${periodoFormateado ? `<span class="info-label">Período:</span> ${periodoFormateado}` : ''}</p>
    </div>`;

    for (let day = 1; day <= 3; day++) {
        const dayTitle = document.getElementById(`day${day}-title`).value || `Día ${day}`;
        const exercises = document.querySelectorAll(`#day${day}-exercises .exercise-row`);

        html += `
    <div class="day-section">
        <div class="day-header">DÍA ${day}: ${dayTitle}</div>
        <table>
            <thead>
                <tr>
                    <th class="col-num">#</th>
                    <th>EJERCICIO</th>
                    <th class="col-sets">SERIES</th>
                    <th class="col-reps">REPS</th>
                    <th class="col-rest">OBS</th>
                </tr>
            </thead>
            <tbody>`;

        exercises.forEach((row, index) => {
            const name = row.querySelector('.exercise-name').value;
            const sets = row.querySelector('.exercise-sets').value;
            const reps = row.querySelector('.exercise-reps').value;
            const rest = row.querySelector('.exercise-obs').value;

            if (name) {
                html += `
                <tr>
                    <td class="col-num">${index + 1}</td>
                    <td>${name}</td>
                    <td class="col-sets">${sets}</td>
                    <td class="col-reps">${reps}</td>
                    <td class="col-rest">${rest}</td>
                </tr>`;
            }
        });

        html += `
            </tbody>
        </table>
    </div>`;
    }

    html += `
</body>
</html>`;

    const printWindow = window.open('', '_blank');
    printWindow.document.write(html);
    printWindow.document.close();
    printWindow.onload = function() {
        printWindow.print();
    };
}