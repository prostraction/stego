document.getElementById('file-input').addEventListener('change', function(e) {
    if (e.target.files && e.target.files[0]) {
        const reader = new FileReader();
        reader.onload = function(e) {
            document.getElementById('original-image').src = e.target.result;
        };
        reader.readAsDataURL(e.target.files[0]);
    }
});

document.getElementById('stegano-form').addEventListener('submit', function(e) {
    e.preventDefault();
    document.getElementById('processed-image').src = '/api/placeholder/400/300';
});

document.getElementsByName('input-type').forEach(radio => {
    radio.addEventListener('change', function(e) {
        const fileInput = document.getElementById('file-input');
        if (e.target.value === 'folder') {
            fileInput.setAttribute('webkitdirectory', '');
            fileInput.setAttribute('directory', '');
        } else {
            fileInput.removeAttribute('webkitdirectory');
            fileInput.removeAttribute('directory');
        }
    });
});