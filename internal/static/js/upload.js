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
        }
    }

    function hideMessages() {
        errorMsg.classList.remove('visible');
        successMsg.classList.remove('visible');
        resultCard.classList.remove('visible');
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
})();
