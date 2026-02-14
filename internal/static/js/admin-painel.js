(function () {
    'use strict';

    var pinSection = document.getElementById('pin-section');
    var pinInput = document.getElementById('pin-input');
    var pinBtn = document.getElementById('pin-btn');
    var errorMessage = document.getElementById('error-message');
    var spinner = document.getElementById('spinner');
    var dataSection = document.getElementById('data-section');
    var researchersBody = document.getElementById('researchers-body');
    var totalCount = document.getElementById('total-count');

    function showError(msg) {
        errorMessage.textContent = msg;
        errorMessage.style.display = 'block';
    }

    function hideError() {
        errorMessage.style.display = 'none';
    }

    function setLoading(on) {
        spinner.style.display = on ? 'block' : 'none';
    }

    function escapeHtml(text) {
        var div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    function loadResearchers(pin) {
        hideError();
        setLoading(true);

        fetch('/api/admin/researchers', {
            method: 'GET',
            headers: { 'X-Admin-PIN': pin }
        })
        .then(function (res) { return res.json(); })
        .then(function (data) {
            setLoading(false);
            if (!data.success) {
                showError(data.error || 'Erro desconhecido');
                return;
            }

            sessionStorage.setItem('adminPIN', pin);
            pinSection.style.display = 'none';
            dataSection.style.display = 'block';

            var researchers = data.researchers || [];
            researchersBody.innerHTML = '';

            researchers.forEach(function (r) {
                var tr = document.createElement('tr');

                // Name linked to Lattes
                var tdName = document.createElement('td');
                var nameLink = document.createElement('a');
                nameLink.href = 'https://lattes.cnpq.br/' + encodeURIComponent(r.lattesId);
                nameLink.target = '_blank';
                nameLink.rel = 'noopener noreferrer';
                nameLink.textContent = r.name;
                tdName.appendChild(nameLink);
                tr.appendChild(tdName);

                // Resumo
                var tdResumo = document.createElement('td');
                if (r.hasResumo) {
                    var resumoLink = document.createElement('a');
                    resumoLink.href = '/?resumo=' + encodeURIComponent(r.lattesId);
                    resumoLink.target = '_blank';
                    resumoLink.rel = 'noopener noreferrer';
                    resumoLink.textContent = 'Ver resumo';
                    tdResumo.appendChild(resumoLink);
                } else {
                    tdResumo.textContent = '\u2014';
                    tdResumo.style.color = 'var(--color-text-muted)';
                }
                tr.appendChild(tdResumo);

                // Análise
                var tdAnalise = document.createElement('td');
                if (r.hasAnalise) {
                    var analiseLink = document.createElement('a');
                    analiseLink.href = '/?analise=' + encodeURIComponent(r.lattesId);
                    analiseLink.target = '_blank';
                    analiseLink.rel = 'noopener noreferrer';
                    analiseLink.textContent = 'Ver an\u00e1lise';
                    tdAnalise.appendChild(analiseLink);
                } else {
                    tdAnalise.textContent = '\u2014';
                    tdAnalise.style.color = 'var(--color-text-muted)';
                }
                tr.appendChild(tdAnalise);

                researchersBody.appendChild(tr);
            });

            totalCount.textContent = 'Total: ' + researchers.length + ' pesquisador' + (researchers.length !== 1 ? 'es' : '');
        })
        .catch(function () {
            setLoading(false);
            showError('Erro ao conectar com o servidor');
        });
    }

    pinBtn.addEventListener('click', function () {
        var pin = pinInput.value.trim();
        if (!pin) {
            showError('Digite o PIN');
            return;
        }
        loadResearchers(pin);
    });

    pinInput.addEventListener('keydown', function (e) {
        if (e.key === 'Enter') {
            pinBtn.click();
        }
    });

    // Auto-login if PIN in sessionStorage
    var savedPIN = sessionStorage.getItem('adminPIN');
    if (savedPIN) {
        loadResearchers(savedPIN);
    }
})();
