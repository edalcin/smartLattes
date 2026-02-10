(function () {
    var form = document.getElementById('upload-form');
    var dropZone = document.getElementById('drop-zone');
    var fileInput = document.getElementById('file-input');
    var fileNameDisplay = document.getElementById('file-name');
    var submitBtn = document.getElementById('submit-btn');
    var spinner = document.getElementById('spinner');
    var errorMsg = document.getElementById('error-message');
    var successMsg = document.getElementById('success-message');
    var resultCard = document.getElementById('result-card');

    // AI elements
    var aiSection = document.getElementById('ai-section');
    var providerSelect = document.getElementById('provider-select');
    var apiKeyInput = document.getElementById('api-key-input');
    var loadModelsBtn = document.getElementById('load-models-btn');
    var modelSelect = document.getElementById('model-select');
    var generateBtn = document.getElementById('generate-btn');
    var aiSpinner = document.getElementById('ai-spinner');
    var aiLoadingMsg = document.getElementById('ai-loading-message');
    var aiError = document.getElementById('ai-error');
    var summarySection = document.getElementById('summary-section');
    var truncationWarning = document.getElementById('truncation-warning');
    var summaryContent = document.getElementById('summary-content');
    var downloadMd = document.getElementById('download-md');
    var saveBtn = document.getElementById('save-btn');

    // Analysis elements
    var analysisPromptSection = document.getElementById('analysis-prompt-section');
    var analysisYesBtn = document.getElementById('analysis-yes-btn');
    var analysisNoBtn = document.getElementById('analysis-no-btn');
    var analysisSection = document.getElementById('analysis-section');
    var analysisSpinner = document.getElementById('analysis-spinner');
    var analysisLoadingMsg = document.getElementById('analysis-loading-message');
    var analysisError = document.getElementById('analysis-error');
    var analysisInfo = document.getElementById('analysis-info');
    var analysisResult = document.getElementById('analysis-result');
    var analysisTruncationWarning = document.getElementById('analysis-truncation-warning');
    var analysisContent = document.getElementById('analysis-content');
    var analysisDownloadMd = document.getElementById('analysis-download-md');
    var analysisSaveBtn = document.getElementById('analysis-save-btn');
    var currentAnalysis = '';
    var currentResearchersAnalyzed = 0;

    var currentLattesId = '';
    var currentSummary = '';
    var currentProvider = '';
    var currentModel = '';

    dropZone.addEventListener('click', function () {
        fileInput.click();
    });

    dropZone.addEventListener('dragover', function (e) {
        e.preventDefault();
        dropZone.classList.add('dragover');
    });

    dropZone.addEventListener('dragleave', function () {
        dropZone.classList.remove('dragover');
    });

    dropZone.addEventListener('drop', function (e) {
        e.preventDefault();
        dropZone.classList.remove('dragover');

        var files = e.dataTransfer.files;
        if (files.length > 0) {
            handleFile(files[0]);
        }
    });

    fileInput.addEventListener('change', function () {
        if (fileInput.files.length > 0) {
            handleFile(fileInput.files[0]);
        }
    });

    function handleFile(file) {
        if (!file.name.toLowerCase().endsWith('.xml')) {
            showError('Por favor, selecione um arquivo .xml');
            submitBtn.disabled = true;
            fileNameDisplay.textContent = '';
            return;
        }

        fileNameDisplay.textContent = file.name;
        submitBtn.disabled = false;
        hideMessages();
    }

    form.addEventListener('submit', function (e) {
        e.preventDefault();

        var file = fileInput.files[0];
        if (!file) {
            showError('Selecione um arquivo para enviar');
            return;
        }

        hideMessages();
        spinner.classList.add('visible');
        submitBtn.disabled = true;

        var formData = new FormData();
        formData.append('file', file);

        fetch('/api/upload', {
            method: 'POST',
            body: formData,
        })
            .then(function (response) {
                return response.json().then(function (data) {
                    return { status: response.status, body: data };
                });
            })
            .then(function (result) {
                spinner.classList.remove('visible');

                if (result.body.success) {
                    showSuccess(result.body);
                } else {
                    showError(result.body.error || 'Erro desconhecido ao processar o arquivo');
                }

                submitBtn.disabled = false;
            })
            .catch(function () {
                spinner.classList.remove('visible');
                showError('Erro de conexão. Verifique se o servidor está disponível.');
                submitBtn.disabled = false;
            });
    });

    function showError(message) {
        hideMessages();
        errorMsg.textContent = message;
        errorMsg.classList.add('visible');
    }

    function showSuccess(data) {
        hideMessages();

        var msg = data.updated ? 'CV atualizado com sucesso' : 'CV importado com sucesso';
        successMsg.textContent = msg;
        successMsg.classList.add('visible');

        if (data.data) {
            document.getElementById('result-name').textContent = data.data.name || '-';
            document.getElementById('result-lattes-id').textContent = data.data.lattesId || '-';
            document.getElementById('result-update').textContent = formatDate(data.data.lastUpdate);

            if (data.data.counts) {
                document.getElementById('count-biblio').textContent = data.data.counts.bibliographicProduction || 0;
                document.getElementById('count-tech').textContent = data.data.counts.technicalProduction || 0;
                document.getElementById('count-other').textContent = data.data.counts.otherProduction || 0;
            }

            resultCard.classList.add('visible');

            // Show AI section and store lattesId
            currentLattesId = data.data.lattesId;
            if (aiSection) {
                aiSection.style.display = 'block';
            }
        }
    }

    function hideMessages() {
        errorMsg.classList.remove('visible');
        successMsg.classList.remove('visible');
        resultCard.classList.remove('visible');
        if (aiSection) aiSection.style.display = 'none';
        if (summarySection) summarySection.style.display = 'none';
        if (analysisPromptSection) analysisPromptSection.style.display = 'none';
        if (analysisSection) analysisSection.style.display = 'none';
    }

    function formatDate(dateStr) {
        if (!dateStr || dateStr.length !== 8) {
            return dateStr || '-';
        }
        var day = dateStr.substring(0, 2);
        var month = dateStr.substring(2, 4);
        var year = dateStr.substring(4, 8);
        return day + '/' + month + '/' + year;
    }

    // AI Section handlers (only if elements exist)
    if (providerSelect && apiKeyInput && loadModelsBtn) {
        providerSelect.addEventListener('change', checkLoadModels);
        apiKeyInput.addEventListener('input', checkLoadModels);

        function checkLoadModels() {
            loadModelsBtn.disabled = !(providerSelect.value && apiKeyInput.value.length >= 10);
        }

        loadModelsBtn.addEventListener('click', function () {
            hideAiError();
            loadModelsBtn.disabled = true;
            aiSpinner.classList.add('visible');

            fetch('/api/models', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    provider: providerSelect.value,
                    apiKey: apiKeyInput.value
                })
            })
            .then(function (r) { return r.json(); })
            .then(function (data) {
                aiSpinner.classList.remove('visible');
                loadModelsBtn.disabled = false;

                if (!data.success) {
                    showAiError(data.error || 'Erro ao carregar modelos');
                    return;
                }

                modelSelect.innerHTML = '<option value="">Selecione o modelo...</option>';
                for (var i = 0; i < data.models.length; i++) {
                    var opt = document.createElement('option');
                    opt.value = data.models[i].id;
                    opt.textContent = data.models[i].displayName || data.models[i].id;
                    modelSelect.appendChild(opt);
                }
                modelSelect.disabled = false;
            })
            .catch(function () {
                aiSpinner.classList.remove('visible');
                loadModelsBtn.disabled = false;
                showAiError('Erro de conexão ao carregar modelos');
            });
        });

        modelSelect.addEventListener('change', function () {
            generateBtn.disabled = !modelSelect.value;
        });

        generateBtn.addEventListener('click', function () {
            hideAiError();
            generateBtn.disabled = true;
            aiSpinner.classList.add('visible');
            aiLoadingMsg.style.display = 'block';
            summarySection.style.display = 'none';

            currentProvider = providerSelect.value;
            currentModel = modelSelect.value;

            fetch('/api/summary', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    lattesId: currentLattesId,
                    provider: currentProvider,
                    apiKey: apiKeyInput.value,
                    model: currentModel
                })
            })
            .then(function (r) { return r.json(); })
            .then(function (data) {
                aiSpinner.classList.remove('visible');
                aiLoadingMsg.style.display = 'none';
                generateBtn.disabled = false;

                if (!data.success) {
                    showAiError(data.error || 'Erro ao gerar resumo');
                    return;
                }

                currentSummary = data.summary;

                if (data.truncated) {
                    truncationWarning.textContent = data.truncationWarning;
                    truncationWarning.style.display = 'block';
                } else {
                    truncationWarning.style.display = 'none';
                }

                summaryContent.innerHTML = renderMarkdown(data.summary);
                summarySection.style.display = 'block';

                if (analysisPromptSection) analysisPromptSection.style.display = 'block';
            })
            .catch(function () {
                aiSpinner.classList.remove('visible');
                aiLoadingMsg.style.display = 'none';
                generateBtn.disabled = false;
                showAiError('Erro de conexão ao gerar resumo');
            });
        });

        downloadMd.addEventListener('click', function () {
            downloadBlob(currentSummary, 'resumo-' + currentLattesId + '.md', 'text/markdown');
            saveSummary();
        });
        saveBtn.addEventListener('click', function () {
            saveSummary();
        });
    }

    function saveSummary() {
        fetch('/api/summary/save', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                lattesId: currentLattesId,
                summary: currentSummary,
                provider: currentProvider,
                model: currentModel
            })
        })
        .then(function (r) { return r.json(); })
        .then(function (data) {
            if (data.success && saveBtn) {
                saveBtn.textContent = 'Salvo!';
                setTimeout(function () { saveBtn.textContent = 'Ok'; }, 2000);
            }
        })
        .catch(function () { });
    }

    // Analysis handlers
    if (analysisYesBtn) {
        analysisYesBtn.addEventListener('click', function () {
            analysisPromptSection.style.display = 'none';
            analysisSection.style.display = 'block';
            analysisSpinner.classList.add('visible');
            analysisLoadingMsg.style.display = 'block';
            analysisError.classList.remove('visible');
            analysisInfo.classList.remove('visible');
            analysisResult.style.display = 'none';

            fetch('/api/analysis', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    lattesId: currentLattesId,
                    provider: currentProvider,
                    apiKey: apiKeyInput.value,
                    model: currentModel
                })
            })
            .then(function (r) {
                return r.json().then(function (data) {
                    return { status: r.status, body: data };
                });
            })
            .then(function (result) {
                analysisSpinner.classList.remove('visible');
                analysisLoadingMsg.style.display = 'none';

                if (result.status === 409) {
                    analysisInfo.textContent = result.body.error || 'Não há outros pesquisadores para analisar.';
                    analysisInfo.classList.add('visible');
                    return;
                }

                if (!result.body.success) {
                    analysisError.textContent = result.body.error || 'Erro ao analisar relações';
                    analysisError.classList.add('visible');
                    return;
                }

                currentAnalysis = result.body.analysis;
                currentResearchersAnalyzed = result.body.researchersAnalyzed || 0;

                if (result.body.truncated) {
                    analysisTruncationWarning.textContent = result.body.truncationWarning;
                    analysisTruncationWarning.style.display = 'block';
                } else {
                    analysisTruncationWarning.style.display = 'none';
                }

                analysisContent.innerHTML = renderMarkdown(result.body.analysis);
                analysisResult.style.display = 'block';
            })
            .catch(function () {
                analysisSpinner.classList.remove('visible');
                analysisLoadingMsg.style.display = 'none';
                analysisError.textContent = 'Erro de conexão ao analisar relações';
                analysisError.classList.add('visible');
            });
        });
    }

    if (analysisNoBtn) {
        analysisNoBtn.addEventListener('click', function () {
            analysisPromptSection.style.display = 'none';
        });
    }

    if (analysisDownloadMd) {
        analysisDownloadMd.addEventListener('click', function () {
            downloadBlob(currentAnalysis, 'analise-' + currentLattesId + '.md', 'text/markdown');
            saveAnalysis();
        });
    }
    if (analysisSaveBtn) {
        analysisSaveBtn.addEventListener('click', function () {
            saveAnalysis();
        });
    }

    function saveAnalysis() {
        fetch('/api/analysis/save', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                lattesId: currentLattesId,
                analysis: currentAnalysis,
                provider: currentProvider,
                model: currentModel,
                researchersAnalyzed: currentResearchersAnalyzed
            })
        })
        .then(function (r) { return r.json(); })
        .then(function (data) {
            if (data.success && analysisSaveBtn) {
                analysisSaveBtn.textContent = 'Salvo!';
                setTimeout(function () { analysisSaveBtn.textContent = 'Ok'; }, 2000);
            }
        })
        .catch(function () { });
    }

    function showAiError(message) {
        if (aiError) {
            aiError.textContent = message;
            aiError.classList.add('visible');
        }
    }

    function hideAiError() {
        if (aiError) {
            aiError.classList.remove('visible');
        }
    }

    function downloadBlob(content, filename, mimeType) {
        var blob = new Blob([content], { type: mimeType + '; charset=utf-8' });
        var url = URL.createObjectURL(blob);
        var a = document.createElement('a');
        a.href = url;
        a.download = filename;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
    }

    function renderMarkdown(md) {
        var html = md
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;');

        html = html.replace(/^### (.+)$/gm, '<h4>$1</h4>');
        html = html.replace(/^## (.+)$/gm, '<h3>$1</h3>');
        html = html.replace(/^# (.+)$/gm, '<h2>$1</h2>');
        html = html.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>');
        html = html.replace(/\*(.+?)\*/g, '<em>$1</em>');

        html = html.replace(/^\|(.+)\|$/gm, function (match, content) {
            var cells = content.split('|').map(function (c) { return c.trim(); });
            return '<tr>' + cells.map(function (c) {
                if (/^[-:]+$/.test(c)) return '';
                return '<td>' + c + '</td>';
            }).join('') + '</tr>';
        });
        html = html.replace(/((?:<tr>.*?<\/tr>\n?)+)/g, '<table class="summary-table">$1</table>');
        html = html.replace(/<tr><\/tr>/g, '');

        html = html.replace(/^- (.+)$/gm, '<li>$1</li>');
        html = html.replace(/((?:<li>.*?<\/li>\n?)+)/g, '<ul>$1</ul>');

        html = html.replace(/^(?!<[hultd])(.+)$/gm, '<p>$1</p>');
        html = html.replace(/<p>\s*<\/p>/g, '');

        return html;
    }
})();
