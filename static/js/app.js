const API_BASE = 'http://localhost:8090/v1';

const state = {
    user: JSON.parse(localStorage.getItem('user')) || null,
    token: localStorage.getItem('token') || null,
    cart: JSON.parse(localStorage.getItem('cart')) || [],
    products: [],
    view: 'auth', // 'auth' | 'dashboard'
    authMode: 'login' // 'login' | 'register'
};

const app = document.getElementById('app');
const toastContainer = document.getElementById('toast-container');

document.addEventListener('DOMContentLoaded', () => {
    const urlParams = new URLSearchParams(window.location.search);
    const token = urlParams.get('token');
    const paymentId = urlParams.get('paymentId');
    const subscriptionId = urlParams.get('subscription_id');

    if (token || paymentId || subscriptionId) {
        if (state.token) {
            handlePayPalReturn(urlParams);
        } else {
            const tempUser = JSON.parse(localStorage.getItem('temp_user'));
            if (tempUser) {
                state.user = tempUser.user;
                state.token = tempUser.token;
                handlePayPalReturn(urlParams);
            }
        }
    }

    render();
});

function render() {
    app.innerHTML = '';
    if (state.token && state.user) {
        renderDashboard();
    } else {
        renderAuth();
    }
}

function renderAuth() {
    const container = document.createElement('div');
    container.className = 'auth-container view-transition';

    const card = document.createElement('div');
    card.className = 'auth-card';

    // Tabs
    const tabs = document.createElement('div');
    tabs.className = 'auth-tabs';
    tabs.innerHTML = `
        <div class="auth-tab ${state.authMode === 'login' ? 'active' : ''}" onclick="setAuthMode('login')">Login</div>
        <div class="auth-tab ${state.authMode === 'register' ? 'active' : ''}" onclick="setAuthMode('register')">Register</div>
    `;

    // Form
    const form = document.createElement('form');
    form.onsubmit = handleAuthSubmit;

    let formContent = '';
    if (state.authMode === 'register') {
        formContent += `
            <div class="form-group">
                <label class="form-label">Full Name</label>
                <input type="text" name="name" class="form-control" placeholder="John Doe" required>
            </div>
        `;
    }

    formContent += `
        <div class="form-group">
            <label class="form-label">Email Address</label>
            <input type="email" name="email" class="form-control" placeholder="name@example.com" required>
        </div>
        <div class="form-group">
            <label class="form-label">Password</label>
            <input type="password" name="password" class="form-control" placeholder="••••••••" required>
        </div>
        <button type="submit" class="btn btn-primary btn-block">
            ${state.authMode === 'login' ? 'Sign In' : 'Create Account'}
        </button>
    `;

    form.innerHTML = formContent;

    // Demo Hint
    const hint = document.createElement('p');
    hint.style.fontSize = '0.8rem';
    hint.style.color = '#6c757d';
    hint.style.marginTop = '1rem';
    hint.style.textAlign = 'center';
    hint.innerHTML = 'Demo Mode: Use any email/password to register.';

    card.appendChild(tabs);
    card.appendChild(form);
    card.appendChild(hint);
    container.appendChild(card);
    app.appendChild(container);
}

function setAuthMode(mode) {
    state.authMode = mode;
    render();
}

async function handleAuthSubmit(e) {
    e.preventDefault();
    const formData = new FormData(e.target);
    const data = Object.fromEntries(formData.entries());

    try {
        let endpoint = state.authMode === 'login' ? '/auth/login' : '/auth/register';

        if (state.authMode === 'register' && !data.name) {
            data.name = data.email.split('@')[0];
        }

        const payload = state.authMode === 'register'
            ? { FullName: data.name, Email: data.email, Password: data.password }
            : { Email: data.email, Password: data.password };

        const res = await apiCall(endpoint, 'POST', payload);

        if (state.authMode === 'register') {
            showToast('Registration successful! Please login.', 'success');
            setAuthMode('login');
        } else {
            if (res.data && res.data.access_token) {
                state.token = res.data.access_token;
                state.user = {
                    email: data.email,
                    id: parseJwt(res.data.access_token).data.user_id
                };

                saveState();
                showToast('Welcome back!', 'success');
                render();
                loadDashboardData();
            }
        }
    } catch (err) {
        showToast(err.message || 'Authentication failed', 'error');
    }
}


function renderDashboard() {
    const header = document.createElement('header');
    header.className = 'main-header';
    header.innerHTML = `
        <div class="container header-content">
            <a href="#" class="logo">
                <span style="font-size: 1.2em;"></span> Demo
            </a>
            <div class="nav-actions">
                <div class="user-profile">
                    <span></span> ${state.user.email}
                </div>
                <button class="btn btn-sm btn-outline" onclick="logout()">Logout</button>
                <button class="cart-trigger" onclick="toggleCart()">
                    🛒 <span class="badge" id="cart-badge">${state.cart.reduce((a, b) => a + b.quantity, 0)}</span>
                </button>
            </div>
        </div>
    `;

    // Main Content
    const main = document.createElement('main');
    main.className = 'container view-transition';
    main.style.paddingTop = '2rem';
    main.style.paddingBottom = '3rem';

    const vaultSection = document.createElement('section');
    vaultSection.id = 'vault-section';
    vaultSection.style.marginBottom = '2rem';
    vaultSection.style.display = 'none';
    vaultSection.innerHTML = `
        <h3>Saved Payment Methods</h3>
        <div class="grid" id="vault-grid" style="grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));"></div>
    `;

    const productSection = document.createElement('section');
    productSection.innerHTML = `
        <h3>Products</h3>
        <div class="grid" id="product-grid">
            <div class="card"><p>Loading products...</p></div>
        </div>
    `;

    const historySection = document.createElement('section');
    historySection.style.marginTop = '3rem';
    historySection.innerHTML = `
        <h3>Order History</h3>
        <div id="order-history" style="margin-top:1rem;">
            <p>Loading history...</p>
        </div>
    `;

    main.appendChild(vaultSection);
    main.appendChild(productSection);
    main.appendChild(historySection);

    const cartSidebar = document.createElement('div');
    cartSidebar.id = 'cart-sidebar';
    cartSidebar.className = 'cart-sidebar';
    renderCartContent(cartSidebar);

    const overlay = document.createElement('div');
    overlay.id = 'cart-overlay';
    overlay.style.cssText = 'position:fixed;top:0;left:0;width:100%;height:100%;background:rgba(0,0,0,0.5);z-index:900;display:none;opacity:0;transition:opacity 0.3s;';
    overlay.onclick = toggleCart;

    app.appendChild(header);
    app.appendChild(main);
    app.appendChild(cartSidebar);
    app.appendChild(overlay);

    loadDashboardData();
}

async function loadDashboardData() {
    await loadProducts();
    await loadVaultedMethods();
    await loadOrders();
}

async function loadProducts() {
    try {
        const res = await apiCall('/products', 'GET');
        state.products = res.data || [];

        const grid = document.getElementById('product-grid');
        if (!grid) return;

        grid.innerHTML = state.products.map(p => {
            const hasSub = p.pricing_types && p.pricing_types.includes('SUBSCRIPTION') && p.subscription_options && p.subscription_options.length > 0;
            const hasOneTime = p.pricing_types && p.pricing_types.includes('ONE_TIME') && p.one_time_price;

            let pricingOptionsHTML = '';
            let defaultPrice = 0;
            let defaultType = '';

            if (hasSub && hasOneTime) {
                const subOpt = p.subscription_options[0];
                defaultPrice = subOpt.final_price;
                defaultType = 'SUBSCRIPTION';
                pricingOptionsHTML = `
                    <div style="margin-bottom: 0.8rem; display: flex; gap: 0.5rem; flex-wrap: wrap;" id="ptype-container-${p.product.id}">
                        <button class="btn btn-sm btn-primary active" style="flex:1; border-radius: 4px;" data-price="${subOpt.final_price}" onclick="updatePriceDisplay(${p.product.id}, 'SUBSCRIPTION', this)">
                            Sub ($${subOpt.final_price}/${subOpt.interval_unit.substring(0, 2).toLowerCase()})
                        </button>
                        <button class="btn btn-sm btn-outline" style="flex:1; border-radius: 4px;" data-price="${p.one_time_price.price}" onclick="updatePriceDisplay(${p.product.id}, 'ONE_TIME', this)">
                            One-Time ($${p.one_time_price.price})
                        </button>
                    </div>
                `;
            } else if (hasSub) {
                const subOpt = p.subscription_options[0];
                defaultPrice = subOpt.final_price;
                defaultType = 'SUBSCRIPTION';
                let tags = `<span class="product-type sub" style="display:inline-block; margin-bottom: 0.3rem;">Subscription</span>`;
                if (subOpt.has_trial) {
                    tags += ` <span style="font-size: 0.7em; background: #e0f7fa; color: #00796b; padding: 2px 6px; border-radius: 4px; border: 1px solid #b2ebf2; display:inline-block;">Trial: ${subOpt.trial_cycles}x ${subOpt.trial_interval_count} ${subOpt.trial_interval_unit}</span>`;
                }
                pricingOptionsHTML = tags;
            } else if (hasOneTime) {
                defaultPrice = p.one_time_price.price;
                defaultType = 'ONE_TIME';
                pricingOptionsHTML = `<span class="product-type onetime">One-time</span>`;
            }

            return `
                <div class="card">
                    <div style="flex:1;">
                        <input type="hidden" id="selected-type-${p.product.id}" value="${defaultType}" />
                        ${pricingOptionsHTML}
                        <h3 style="margin-top:0.5rem; font-size: 1.2rem;">${p.product.name}</h3>
                        <p style="color:var(--text-muted); font-size:0.9rem;">${p.product.description}</p>
                    </div>
                    <div style="margin-top: 1rem; display:flex; justify-content:space-between; align-items:center;">
                        <span class="product-price" id="price-display-${p.product.id}">$${defaultPrice}</span>
                        <button class="btn btn-primary btn-sm" onclick="addToCart(${p.product.id})">
                            Add to Cart
                        </button>
                    </div>
                </div>
            `;
        }).join('');
    } catch (err) {
        console.error(err);
        showToast('Failed to load products', 'error');
    }
}

function updatePriceDisplay(productId, type, btnElement) {
    if (btnElement) {
        const container = document.getElementById('ptype-container-' + productId);
        if (container) {
            const buttons = container.querySelectorAll('button');
            buttons.forEach(b => {
                b.classList.remove('active', 'btn-primary');
                b.classList.add('btn-outline');
            });
            btnElement.classList.remove('btn-outline');
            btnElement.classList.add('active', 'btn-primary');
        }

        const price = btnElement.getAttribute('data-price');
        const priceDisplay = document.getElementById('price-display-' + productId);
        if (priceDisplay) {
            priceDisplay.innerText = '$' + price;
        }
    }

    const typeInput = document.getElementById('selected-type-' + productId);
    if (typeInput) {
        typeInput.value = type;
    }
}

async function loadOrders() {
    try {
        const res = await apiCall('/orders', 'GET');
        const container = document.getElementById('order-history');
        if (!container) return;

        if (!res.data || res.data.length === 0) {
            container.innerHTML = '<p class="text-center" style="color:var(--text-muted);">No orders found.</p>';
            return;
        }

        container.innerHTML = `<div class="grid" style="grid-template-columns: repeat(auto-fill, minmax(350px, 1fr)); gap: 1.5rem;">
            ${res.data.map(o => renderOrderCard(o)).join('')}
        </div>`;

    } catch (err) {
        console.error('Failed to load orders', err);
        const container = document.getElementById('order-history');
        if (container) container.innerHTML = '<p class="text-center" style="color:var(--danger);">Failed to load order history.</p>';
    }
}

function renderOrderCard(o) {
    const isSub = o.order_type === 'SUBSCRIPTION' && o.subscription;
    const statusColor = getStatusColor(o.status);
    const date = new Date(o.created_at).toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });

    let content = '';

    if (isSub) {
        const sub = o.subscription;

        // Build extra details if they exist
        const nextBillingHTML = sub.next_billing_time ? `
            <div style="display:flex; justify-content:space-between; margin-bottom: 0.3rem;">
                <span>Next Billing Time:</span>
                <span>${new Date(sub.next_billing_time).toLocaleDateString()}</span>
            </div>
        ` : '';

        const lastPaymentHTML = sub.last_payment_time ? `
            <div style="display:flex; justify-content:space-between; margin-bottom: 0.3rem;">
                <span>Last Payment Time:</span>
                <span>${new Date(sub.last_payment_time).toLocaleDateString()}</span>
            </div>
        ` : '';

        content = `
            <div style="margin-bottom: 1rem;">
                <div style="font-size: 0.85rem; text-transform: uppercase; letter-spacing: 0.5px; color: var(--secondary); font-weight: 700;">
                    Subscription
                </div>
                <!-- <div style="margin: 0.5rem 0;">
                    <span style="font-size: 0.75rem; text-transform: uppercase; letter-spacing: 0.5px; color: ${statusColor}; font-weight: 700; border: 1px solid ${statusColor}; padding: 0.2em 0.5em; border-radius: 4px;">${o.status}</span>
                </div> -->
                <h4 style="margin: 0.5rem 0; font-size: 1.1rem;">${sub.plan_name || 'Subscription Plan'}</h4>
                <div style="font-size: 0.9rem; color: var(--text-muted);">
                    ${sub.price_value} ${sub.currency} / ${sub.interval_count} ${sub.interval_unit}
                </div>
            </div>
            <div style="background: var(--bg-body); padding: 0.8rem; border-radius: var(--radius-sm); font-size: 0.85rem;">
                <div style="display:flex; justify-content:space-between; margin-bottom: 0.3rem;">
                    <span>Subscription ID:</span>
                    <span style="font-family:monospace;">${sub.paypal_subscription_id || '-'}</span>
                </div>
                <div style="display:flex; justify-content:space-between; margin-bottom: 0.3rem;">
                    <span>Start Date:</span>
                    <span>${sub.start_time ? new Date(sub.start_time).toLocaleDateString() : '-'}</span>
                </div>
                ${nextBillingHTML}
                ${lastPaymentHTML}
                <div style="display:flex; justify-content:space-between; margin-bottom: 0.3rem;">
                    <span>Interval Unit:</span>
                    <span>${sub.interval_unit || '-'}</span>
                </div>
                <div style="display:flex; justify-content:space-between; margin-bottom: 0.3rem;">
                    <span>Interval Count:</span>
                    <span>${sub.interval_count || '-'}</span>
                </div>
                <div style="display:flex; justify-content:space-between;">
                    <span>Status:</span>
                    <span style="color:${statusColor}; font-weight:700;">${sub.status || '-'}</span>
                </div>
            </div>
        `;
    } else {
        // One Time
        const orderIdDisplay = o.paypal_order_id ? o.paypal_order_id : o.order_id;

        const itemsList = o.items ? o.items.map(i => `
            <div style="display:flex; justify-content:space-between; font-size: 0.9rem; margin-bottom: 0.5rem; border-bottom: 1px dashed var(--border-light); padding-bottom: 0.5rem;">
                <span>${i.quantity}x ${i.product_name}</span>
                <span style="font-weight:600;">$${i.total_price}</span>
            </div>
        `).join('') : '<div style="font-style:italic; color:var(--text-muted);">No items details</div>';

        content = `
             <div style="margin-bottom: 1rem;">
                <div style="font-size: 0.85rem; text-transform: uppercase; letter-spacing: 0.5px; color: var(--primary); font-weight: 700;">
                    One-Time Purchase
                </div>
                <div style="margin: 0.5rem 0;">
                    <span style="font-size: 0.75rem; text-transform: uppercase; letter-spacing: 0.5px; color: ${statusColor}; font-weight: 700; border: 1px solid ${statusColor}; padding: 0.2em 0.5em; border-radius: 4px;">${o.status}</span>
                </div>
                <div style="margin-top: 0.8rem;">
                    ${itemsList}
                </div>
            </div>
             <div style="background: var(--bg-body); padding: 0.8rem; border-radius: var(--radius-sm); font-size: 0.85rem; display:flex; flex-direction:column; gap:0.5rem;">
                <div style="display:flex; justify-content:space-between;">
                    <span>Order ID:</span>
                    <span style="font-family:monospace;">${orderIdDisplay || '-'}</span>
                </div>
                <div style="display:flex; justify-content:space-between; font-weight: 700; border-top: 1px dashed var(--border-light); padding-top: 0.5rem; margin-top: 0.3rem;">
                    <span>Total Paid</span>
                    <span>$${o.total_value} ${o.currency}</span>
                </div>
            </div>
        `;
    }

    return `
        <div class="card" style="border-top: 4px solid ${statusColor}; position: relative;">
            <div style="display: flex; justify-content: space-between; align-items: start; margin-bottom: 1rem; padding-bottom: 0.8rem; border-bottom: 1px solid var(--border-light);">
                <div>
                    <span style="font-weight: 700; font-size: 1.1em;">${date}</span>
                </div>
            </div>
            ${content}
             ${o.status === 'WAITING_PAYMENT' || o.status === 'PENDING' ? `
                <div style="margin-top: 1rem;">
                    <button class="btn btn-sm btn-outline btn-block" onclick="checkout()">Complete Payment</button>
                </div>
            ` : ''}
        </div>
    `;
}

function getStatusColor(status) {
    switch (status) {
        case 'PAID':
        case 'ACTIVE':
        case 'COMPLETED': return 'var(--success)';
        case 'WAITING_PAYMENT':
        case 'PENDING':
        case 'APPROVAL_PENDING': return 'var(--warning)';
        case 'CANCELLED':
        case 'EXPIRED': return 'var(--danger)';
        default: return 'var(--text-muted)';
    }
}

async function loadVaultedMethods() {
    try {
        // Mock endpoint based on previous file, assuming it exists or will exist
        // The user ID is needed. 
        const res = await apiCall(`/payments/vault-tokens`);

        const grid = document.getElementById('vault-grid');
        const section = document.getElementById('vault-section');

        if (res.data && res.data.items && res.data.items.length > 0) {
            section.style.display = 'block';
            grid.innerHTML = res.data.items.map(pm => `
                <div class="payment-method-card">
                    <div>
                        <div style="font-weight:600;">PayPal Saved</div>
                        <div style="font-size:0.85rem; color:var(--text-muted);">${pm.email}</div>
                    </div>
                    <button class="btn btn-outline btn-sm" onclick="payWithVault(${pm.id})">Pay</button>
                </div>
            `).join('');
        } else {
            section.style.display = 'none';
        }
    } catch (err) {
        // Silent fail for vault loading
        console.log('Vault load error', err);
    }
}

// --- Cart Logic ---
function addToCart(productId) {
    const rawProd = state.products.find(p => p.product.id === productId);
    if (!rawProd) return;

    // Get the selected type from the hidden input we created
    const typeInput = document.getElementById(`selected-type-${productId}`);
    const selectedType = typeInput ? typeInput.value : 'ONE_TIME';

    let isSub = false;
    let price = 0;
    let billingPlanId = null;

    if (selectedType === 'SUBSCRIPTION' && rawProd.subscription_options && rawProd.subscription_options.length > 0) {
        isSub = true;
        price = rawProd.subscription_options[0].final_price;
        billingPlanId = rawProd.subscription_options[0].billing_plan_id;
    } else if (selectedType === 'ONE_TIME' && rawProd.one_time_price) {
        price = rawProd.one_time_price.price;
    } else {
        // Fallback just in case
        isSub = rawProd.subscription_options && rawProd.subscription_options.length > 0;
        price = isSub ? rawProd.subscription_options[0].final_price : rawProd.one_time_price.price;
        billingPlanId = isSub ? rawProd.subscription_options[0].billing_plan_id : null;
    }

    const name = rawProd.product.name;

    const existing = state.cart.find(i => i.product_id === productId && i.is_subscription === isSub);

    if (existing) {
        existing.quantity++;
    } else {
        state.cart.push({
            product_id: productId,
            name: name,
            price: price,
            quantity: 1,
            is_subscription: isSub,
            billing_plan_id: billingPlanId
        });
    }

    saveState();
    updateCartUI();
    toggleCart(true);
    showToast('Added to cart', 'success');
}

function updateCartUI() {
    const badge = document.getElementById('cart-badge');
    const sidebar = document.getElementById('cart-sidebar');
    if (badge) badge.innerText = state.cart.reduce((a, b) => a + b.quantity, 0);
    if (sidebar) renderCartContent(sidebar);
}

function renderCartContent(container) {
    const oneTimeItems = state.cart.filter(i => !i.is_subscription);
    const subItems = state.cart.filter(i => i.is_subscription);

    const oneTimeTotal = oneTimeItems.reduce((sum, item) => sum + (item.price * item.quantity), 0);
    const subTotal = subItems.reduce((sum, item) => sum + (item.price * item.quantity), 0);
    const total = oneTimeTotal + subTotal;

    let html = `
        <div class="cart-header">
            <h3>Cart</h3>
            <span style="cursor:pointer; font-size:1.5rem;" onclick="toggleCart()">×</span>
        </div>
        <div class="cart-body">
            ${state.cart.length === 0 ? '<p class="text-center" style="color:var(--text-muted); margin-top:2rem;">Cart is empty.</p>' : ''}
    `;

    // Render One-Time Items
    if (oneTimeItems.length > 0) {
        html += `
            <div class="cart-section">
                <h5 style="color:var(--primary); border-bottom:1px solid var(--border-light); padding-bottom:5px;">One-Time Items</h5>
                ${oneTimeItems.map((item) => renderCartItem(item)).join('')}
                <div class="text-right" style="font-weight:bold; margin-top:5px;">Subtotal: $${oneTimeTotal.toFixed(2)}</div>
                <button class="btn btn-primary btn-block btn-sm mt-2" onclick="checkout('one-time')">
                    Checkout Items ($${oneTimeTotal.toFixed(2)})
                </button>
            </div>
        `;
    }

    // Render Subscription Items
    if (subItems.length > 0) {
        html += `
            <div class="cart-section" style="margin-top: 1.5rem;">
                <h5 style="color:var(--secondary); border-bottom:1px solid var(--border-light); padding-bottom:5px;">Subscriptions</h5>
                ${subItems.map((item) => renderCartItem(item)).join('')}
                <div class="text-right" style="font-weight:bold; margin-top:5px;">Subtotal: $${subTotal.toFixed(2)}</div>
                <button class="btn btn-primary btn-block btn-sm mt-2" style="background:var(--secondary);" onclick="checkout('subscription')">
                    Activate Subscriptions
                </button>
            </div>
        `;
    }

    html += `</div>`; // End cart-body

    // Footer with Grand Total (informational)
    if (state.cart.length > 0) {
        html += `
            <div class="cart-footer">
                <div style="display:flex; justify-content:space-between; font-weight:700;">
                    <span>Grand Total</span>
                    <span>$${total.toFixed(2)}</span>
                </div>
            </div>
        `;
    }

    container.innerHTML = html;
}

function renderCartItem(item) {
    const idx = state.cart.indexOf(item);
    return `
        <div class="cart-item">
            <div>
                <div style="font-weight:600;">${item.name}</div>
                <div style="font-size:0.85rem; color:var(--text-muted);">
                    $${item.price} x ${item.quantity} 
                    ${item.is_subscription ? '<span style="font-size: 0.65rem; text-transform: uppercase; margin-left:5px; letter-spacing: 0.5px; color: #f72585; font-weight: 700; border: 1px solid #f72585; padding: 0.1em 0.4em; border-radius: 4px;">Subscription</span>' : ''}
                </div>
            </div>
            <div style="text-align:right;">
                <div>$${(item.price * item.quantity).toFixed(2)}</div>
                <small style="color:var(--danger); cursor:pointer;" onclick="removeFromCart(${idx})">Remove</small>
            </div>
        </div>
    `;
}

function removeFromCart(idx) {
    state.cart.splice(idx, 1);
    saveState();
    updateCartUI();
}

function toggleCart(forceOpen = false) {
    const sidebar = document.getElementById('cart-sidebar');
    const overlay = document.getElementById('cart-overlay');
    if (!sidebar || !overlay) return;

    const isOpen = sidebar.classList.contains('open');
    if (forceOpen || !isOpen) {
        sidebar.classList.add('open');
        overlay.style.display = 'block';
        setTimeout(() => overlay.style.opacity = '1', 10);
    } else {
        sidebar.classList.remove('open');
        overlay.style.opacity = '0';
        setTimeout(() => overlay.style.display = 'none', 300);
    }
}

// --- Checkout Logic ---
async function checkout(type) {
    if (state.cart.length === 0) return;

    try {
        if (type === 'subscription') {
            // Filter subscription items
            const subItems = state.cart.filter(i => i.is_subscription);
            if (subItems.length === 0) return showToast('No subscription items', 'error');

            // Demo limitation: Process first subscription only (PayPal usually handles 1 sub per flow)
            const subItem = subItems[0];
            const payload = { billing_plan_id: subItem.billing_plan_id, quantity: subItem.quantity };

            showToast('Processing Subscription...', 'info');
            const res = await apiCall('/subscriptions', 'POST', payload);

            if (res.data && res.data.approval_url) {
                localStorage.setItem('pending_checkout_type', 'subscription');
                window.location.href = res.data.approval_url;
            } else {
                throw new Error('No approval URL returned for subscription');
            }

        } else if (type === 'one-time') {
            const oneTimeItems = state.cart.filter(i => !i.is_subscription);
            if (oneTimeItems.length === 0) return showToast('No items to checkout', 'error');

            const payload = {
                items: oneTimeItems.map(i => ({ product_id: i.product_id, quantity: i.quantity }))
            };

            showToast('Creating Order...', 'info');
            const res = await apiCall('/orders', 'POST', payload);

            if (res.data) {
                const orderId = res.data.OrderID;
                if (!orderId) throw new Error('Order creation failed, no ID');

                const payRes = await apiCall(`/orders/${orderId}/pay`, 'POST');
                if (payRes.data && payRes.data.ApprovalURL) {
                    localStorage.setItem('pending_checkout_type', 'one-time');
                    // Store order ID to verify later?
                    window.location.href = payRes.data.ApprovalURL;
                } else {
                    throw new Error('Payment initialization failed');
                }
            }
        } else {
            // Fallback or specific payment method
            // If type is integer, it's likely a vault payment method ID? 
            // Logic kept simple for now based on previous impl.
        }

    } catch (err) {
        console.error(err);
        showToast(err.message || 'Checkout failed', 'error');
    }
}

async function payWithVault(vaultTokenId) {
    try {

        if (state.cart.length === 0) return showToast('Cart empty', 'error');

        const orderPayload = {
            items: state.cart.map(i => ({ product_id: i.product_id, quantity: i.quantity }))
        };
        const orderRes = await apiCall('/orders', 'POST', orderPayload);
        const orderId = orderRes.data.OrderID;


        const payPayload = {
            order_id: parseInt(orderId),
            vault_token_id: parseInt(vaultTokenId)
        };

        showToast('Processing vaulted payment...', 'info');
        const payRes = await apiCall('/payments/pay-again', 'POST', payPayload);

        if (payRes.status_code === 200) {
            showToast('Payment successful!', 'success');
            state.cart = [];
            saveState();
            updateCartUI();
            toggleCart();
        }

    } catch (err) {
        showToast('Vault payment failed', 'error');
    }
}


async function handlePayPalReturn(params) {
    const token = params.get('token');
    const payerId = params.get('PayerID');
    const subId = params.get('subscription_id');

    window.history.replaceState({}, document.title, window.location.pathname);

    const pendingType = localStorage.getItem('pending_checkout_type');

    if (subId) {
        showToast(`Subscription Active! ID: ${subId}`, 'success');
        state.cart = state.cart.filter(i => !i.is_subscription);
        saveState();
        localStorage.removeItem('pending_checkout_type');
        return;
    }

    if (token && payerId) {
        showToast('Payment successful! Processing...', 'success');

        state.cart = state.cart.filter(i => i.is_subscription);
        saveState();
        localStorage.removeItem('pending_checkout_type');
    }
}


async function apiCall(endpoint, method = 'GET', body = null) {
    const headers = { 'Content-Type': 'application/json' };
    if (state.token) headers['Authorization'] = `Bearer ${state.token}`;

    const config = { method, headers };
    if (body) config.body = JSON.stringify(body);

    const response = await fetch(`${API_BASE}${endpoint}`, config);
    const result = await response.json();

    if (!response.ok || (result.status_code && result.status_code >= 400)) {
        throw new Error(result.message || 'API Error');
    }

    return result;
}

function saveState() {
    localStorage.setItem('user', JSON.stringify(state.user));
    localStorage.setItem('token', state.token);
    localStorage.setItem('cart', JSON.stringify(state.cart));

    if (state.user && state.token) {
        localStorage.setItem('temp_user', JSON.stringify({ user: state.user, token: state.token }));
    }
}

function logout() {
    state.user = null;
    state.token = null;
    state.cart = [];
    state.view = 'auth';
    localStorage.clear();
    render();
    showToast('Logged out successfully');
}

function showToast(msg, type = 'info') {
    const t = document.createElement('div');
    t.className = `toast ${type}`;
    t.innerHTML = `<span>${type === 'success' ? '✅' : type === 'error' ? '❌' : 'ℹ️'}</span> ${msg}`;

    let container = document.getElementById('toast-container');
    if (!container) {
        container = document.createElement('div');
        container.id = 'toast-container';
        container.className = 'toast-container';
        document.body.appendChild(container);
    }

    container.appendChild(t);
    setTimeout(() => {
        t.style.opacity = '0';
        setTimeout(() => t.remove(), 300);
    }, 3000);
}

function parseJwt(token) {
    var base64Url = token.split('.')[1];
    var base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
    var jsonPayload = decodeURIComponent(window.atob(base64).split('').map(function (c) {
        return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
    }).join(''));

    return JSON.parse(jsonPayload);
}

window.appState = state;
