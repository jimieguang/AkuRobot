// 测试JavaScript加载
document.addEventListener('DOMContentLoaded', function() {
    console.log('测试页面JavaScript已成功加载！');
    
    // 添加一个简单的交互功能
    const buttons = document.querySelectorAll('.button');
    buttons.forEach(function(button) {
        button.addEventListener('click', function(e) {
            e.preventDefault();
            alert('按钮点击测试成功！');
        });
    });
    
    // 添加表单提交处理
    const form = document.querySelector('form');
    if (form) {
        form.addEventListener('submit', function(e) {
            e.preventDefault();
            const name = document.getElementById('name').value;
            const email = document.getElementById('email').value;
            const message = document.getElementById('message').value;
            
            alert(`表单提交测试成功！\n姓名: ${name}\n邮箱: ${email}\n留言: ${message}`);
        });
    }
    
    // 添加页面加载完成提示
    const header = document.querySelector('.header');
    if (header) {
        const notification = document.createElement('div');
        notification.style.backgroundColor = '#4CAF50';
        notification.style.color = 'white';
        notification.style.padding = '10px';
        notification.style.marginTop = '10px';
        notification.style.borderRadius = '5px';
        notification.textContent = 'JavaScript功能测试成功！';
        header.appendChild(notification);
        
        // 3秒后自动隐藏通知
        setTimeout(function() {
            notification.style.opacity = '0';
            notification.style.transition = 'opacity 1s';
            setTimeout(function() {
                notification.remove();
            }, 1000);
        }, 3000);
    }
}); 