(function () {
    // Elements
    var searchInput = document.getElementById('search-input');
    var searchResults = document.getElementById('search-results');
    var selectedCv = document.getElementById('selected-cv');
    var selectedName = document.getElementById('selected-name');
    var selectedLattesId = document.getElementById('selected-lattes-id');
    var aiConfig = document.getElementById('ai-config');
    var providerSelect = document.getElementById('provider-select');
    var apiKeyInput = document.getElementById('api-key-input');
    var loadModelsBtn = document.getElementById('load-models-btn');
    var modelSelect = document.getElementById('model-select');
    var generateBtn = document.getElementById('generate-btn');
    var spinner = document.getElementById('spinner');
    var loadingMessage = document.getElementById('loading-message');
    var errorMsg = document.getElementById('error-message');
    var summarySection = document.getElementById('summary-section');
    var truncationWarning = document.getElementById('truncation-warning');
    var summaryContent = document.getElementById('summary-content');
    var downloadMd = document.getElementById('download-md');
    var downloadPdf = document.getElementById('download-pdf');
    var saveBtn = document.getElementById('save-btn');
    var selectedNameEl = document.getElementById('selected-name');

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
    var analysisDownloadPdf = document.getElementById('analysis-download-pdf');
    var analysisSaveBtn = document.getElementById('analysis-save-btn');
    var currentAnalysis = '';
    var currentResearchersAnalyzed = 0;

    var shareBtn = document.getElementById('share-btn');
    var analysisShareBtn = document.getElementById('analysis-share-btn');

    var currentLattesId = '';
    var currentSummary = '';
    var currentProvider = '';
    var currentModel = '';
    var searchTimeout = null;

    // Search with debounce
    searchInput.addEventListener('input', function () {
        var query = searchInput.value.trim();
        if (searchTimeout) clearTimeout(searchTimeout);

        if (query.length < 3) {
            searchResults.innerHTML = '';
            return;
        }

        searchTimeout = setTimeout(function () {
            fetch('/api/search?q=' + encodeURIComponent(query))
                .then(function (r) { return r.json(); })
                .then(function (data) {
                    if (!data.success || !data.results || data.results.length === 0) {
                        searchResults.innerHTML = '<p class="search-empty">Nenhum resultado encontrado</p>';
                        return;
                    }
                    var html = '';
                    for (var i = 0; i < data.results.length; i++) {
                        var cv = data.results[i];
                        html += '<div class="search-result-card" data-lattes-id="' + cv.lattesId + '" data-name="' + escapeHtml(cv.name) + '">';
                        html += '<strong>' + escapeHtml(cv.name) + '</strong>';
                        html += '<span class="search-result-id">' + cv.lattesId + '</span>';
                        html += '</div>';
                    }
                    searchResults.innerHTML = html;

                    // Add click handlers
                    var cards = searchResults.querySelectorAll('.search-result-card');
                    for (var j = 0; j < cards.length; j++) {
                        cards[j].addEventListener('click', function () {
                            selectCV(this.getAttribute('data-lattes-id'), this.getAttribute('data-name'));
                        });
                    }
                })
                .catch(function () {
                    searchResults.innerHTML = '<p class="search-empty">Erro ao buscar</p>';
                });
        }, 300);
    });

    function selectCV(lattesId, name) {
        currentLattesId = lattesId;
        selectedName.textContent = name;
        selectedLattesId.textContent = lattesId;
        selectedCv.style.display = 'block';
        aiConfig.style.display = 'block';
        searchResults.innerHTML = '';
        hideError();
        summarySection.style.display = 'none';
        if (analysisPromptSection) analysisPromptSection.style.display = 'none';
        if (analysisSection) analysisSection.style.display = 'none';
    }

    // Enable/disable load models button
    providerSelect.addEventListener('change', checkLoadModels);
    apiKeyInput.addEventListener('input', checkLoadModels);

    function checkLoadModels() {
        loadModelsBtn.disabled = !(providerSelect.value && apiKeyInput.value.length >= 10);
    }

    // Load models
    loadModelsBtn.addEventListener('click', function () {
        hideError();
        loadModelsBtn.disabled = true;
        spinner.classList.add('visible');

        fetch('/api/models', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                provider: providerSelect.value,
                apiKey: apiKeyInput.value
            })
        })
        .then(function (r) {
            return r.json().then(function (data) {
                return { status: r.status, body: data };
            });
        })
        .then(function (result) {
            spinner.classList.remove('visible');
            loadModelsBtn.disabled = false;

            if (!result.body.success) {
                showError(result.body.error || 'Erro ao carregar modelos');
                return;
            }
            var data = result.body;

            // Populate model select
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
            spinner.classList.remove('visible');
            loadModelsBtn.disabled = false;
            showError('Erro de conexão ao carregar modelos');
        });
    });

    // Enable generate button when model selected
    modelSelect.addEventListener('change', function () {
        generateBtn.disabled = !modelSelect.value;
    });

    // Generate summary
    generateBtn.addEventListener('click', function () {
        hideError();
        generateBtn.disabled = true;
        spinner.classList.add('visible');
        loadingMessage.style.display = 'block';
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
        .then(function (r) {
            return r.json().then(function (data) {
                return { status: r.status, body: data };
            });
        })
        .then(function (result) {
            spinner.classList.remove('visible');
            loadingMessage.style.display = 'none';
            generateBtn.disabled = false;

            if (!result.body.success) {
                showError(result.body.error || 'Erro ao gerar resumo');
                return;
            }
            var data = result.body;

            currentSummary = data.summary;

            // Show truncation warning if needed
            if (data.truncated) {
                truncationWarning.textContent = data.truncationWarning;
                truncationWarning.style.display = 'block';
            } else {
                truncationWarning.style.display = 'none';
            }

            // Render summary (basic markdown to HTML)
            summaryContent.innerHTML = renderMarkdown(data.summary);
            summarySection.style.display = 'block';

            // Indicar que foi salvo automaticamente
            if (saveBtn) {
                saveBtn.textContent = 'Salvo automaticamente';
                saveBtn.disabled = true;
            }

            if (shareBtn) shareBtn.style.display = '';

            if (analysisPromptSection) analysisPromptSection.style.display = 'block';
        })
        .catch(function () {
            spinner.classList.remove('visible');
            loadingMessage.style.display = 'none';
            generateBtn.disabled = false;
            showError('Erro de conexão ao gerar resumo');
        });
    });

    // Share buttons
    if (shareBtn) {
        shareBtn.addEventListener('click', function () {
            fetch('/api/config').then(function(r){return r.json()}).then(function(cfg){
                var url = cfg.shareBaseUrl + '?resumo=' + currentLattesId;
                copyToClipboard(url, shareBtn);
            });
        });
    }
    if (analysisShareBtn) {
        analysisShareBtn.addEventListener('click', function () {
            fetch('/api/config').then(function(r){return r.json()}).then(function(cfg){
                var url = cfg.shareBaseUrl + '?analise=' + currentLattesId;
                copyToClipboard(url, analysisShareBtn);
            });
        });
    }

    // Download buttons
    downloadMd.addEventListener('click', function () {
        downloadBlob(currentSummary, 'resumo-' + currentLattesId + '.md', 'text/markdown');
    });
    downloadPdf.addEventListener('click', function () {
        downloadAsPdf(currentSummary);
    });

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

                if (analysisShareBtn) analysisShareBtn.style.display = '';

                // Indicar que foi salvo automaticamente
                if (analysisSaveBtn) {
                    analysisSaveBtn.textContent = 'Salvo automaticamente';
                    analysisSaveBtn.disabled = true;
                }
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
        });
    }
    if (analysisDownloadPdf) {
        analysisDownloadPdf.addEventListener('click', function () {
            downloadAsPdf(currentAnalysis);
        });
    }

    // Helpers
    function showError(message) {
        errorMsg.textContent = message;
        errorMsg.classList.add('visible');
    }

    function hideError() {
        errorMsg.classList.remove('visible');
    }

    function escapeHtml(text) {
        var div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
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

    function markdownToHtml(md) {
        return '<html><head><meta charset="utf-8"><style>body{font-family:Arial,sans-serif;max-width:800px;margin:2rem auto;padding:0 1rem;line-height:1.6}h1,h2,h3{color:#1e3a5f}table{border-collapse:collapse;width:100%;margin:1rem 0}td,th{border:1px solid #ddd;padding:8px;text-align:left}ul{padding-left:1.5rem}</style></head><body>' +
            renderMarkdown(md) +
            '</body></html>';
    }

    function downloadAsPdf(content) {
        var html = markdownToHtml(content);
        var iframe = document.createElement('iframe');
        iframe.style.position = 'fixed';
        iframe.style.right = '0';
        iframe.style.bottom = '0';
        iframe.style.width = '0';
        iframe.style.height = '0';
        iframe.style.border = 'none';
        document.body.appendChild(iframe);
        iframe.contentDocument.write(html);
        iframe.contentDocument.close();
        iframe.contentWindow.onafterprint = function () { document.body.removeChild(iframe); };
        setTimeout(function () { iframe.contentWindow.print(); }, 250);
    }

    function copyToClipboard(text, btn) {
        if (navigator.clipboard && navigator.clipboard.writeText) {
            navigator.clipboard.writeText(text).then(function () {
                showShareFeedback(btn);
            }).catch(function () {
                fallbackCopy(text, btn);
            });
        } else {
            fallbackCopy(text, btn);
        }
    }

    function fallbackCopy(text, btn) {
        var ta = document.createElement('textarea');
        ta.value = text;
        ta.style.position = 'fixed';
        ta.style.left = '-9999px';
        document.body.appendChild(ta);
        ta.select();
        document.execCommand('copy');
        document.body.removeChild(ta);
        showShareFeedback(btn);
    }

    function showShareFeedback(btn) {
        var original = btn.textContent;
        btn.textContent = 'Link copiado!';
        btn.disabled = true;
        setTimeout(function () { btn.textContent = original; btn.disabled = false; }, 2000);
    }

    function renderMarkdown(md) {
        // Basic markdown rendering
        var html = md
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;');

        // Headers
        html = html.replace(/^### (.+)$/gm, '<h4>$1</h4>');
        html = html.replace(/^## (.+)$/gm, '<h3>$1</h3>');
        html = html.replace(/^# (.+)$/gm, '<h2>$1</h2>');

        // Horizontal rule
        html = html.replace(/^---$/gm, '<hr>');

        // Bold and italic
        html = html.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>');
        html = html.replace(/\*(.+?)\*/g, '<em>$1</em>');

        // Links
        html = html.replace(/\[([^\]]+)\]\((https?:\/\/[^)]+)\)/g, '<a href="$2" target="_blank">$2</a>');

        // Tables (basic support)
        html = html.replace(/^\|(.+)\|$/gm, function (match, content) {
            var cells = content.split('|').map(function (c) { return c.trim(); });
            return '<tr>' + cells.map(function (c) {
                if (/^[-:]+$/.test(c)) return '';
                return '<td>' + c + '</td>';
            }).join('') + '</tr>';
        });
        // Wrap consecutive <tr> in <table>
        html = html.replace(/((?:<tr>.*?<\/tr>\n?)+)/g, '<table class="summary-table">$1</table>');
        // Remove separator rows
        html = html.replace(/<tr><\/tr>/g, '');

        // Lists
        html = html.replace(/^- (.+)$/gm, '<li>$1</li>');
        html = html.replace(/((?:<li>.*?<\/li>\n?)+)/g, '<ul>$1</ul>');

        // Paragraphs (lines not already wrapped)
        html = html.replace(/^(?!<[hultd])(.+)$/gm, '<p>$1</p>');

        // Clean up empty paragraphs
        html = html.replace(/<p>\s*<\/p>/g, '');

        return html;
    }
})();
