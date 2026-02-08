let count = 0;

document.getElementById('increment-btn').addEventListener('click', function() {
    count += 5;
    document.getElementById('counter').textContent = count;
});

document.getElementById('reset-btn').addEventListener('click', function() {
    count = 0;
    document.getElementById('counter').textContent = count;
});