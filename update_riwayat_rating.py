import os

file_path = 'riwayat_pemesanan.html'
with open(file_path, 'r', encoding='utf-8') as f:
    content = f.read()

# 1. Add <div id="sales-rating-header"></div>
header_old = '''            <!-- Delete Button -->
            <button onclick="clearHistory()" class="w-full sm:w-auto flex items-center justify-center gap-2 text-xs font-bold text-red-600 bg-red-50 hover:bg-red-100 px-4 py-2.5 rounded-xl border border-red-100 transition-all active:scale-95">
                <i class="fa-solid fa-trash-can"></i> Hapus Riwayat
            </button>
        </div>

        <!-- Daftar Pesanan -->'''
        
header_new = '''            <!-- Delete Button -->
            <button onclick="clearHistory()" class="w-full sm:w-auto flex items-center justify-center gap-2 text-xs font-bold text-red-600 bg-red-50 hover:bg-red-100 px-4 py-2.5 rounded-xl border border-red-100 transition-all active:scale-95">
                <i class="fa-solid fa-trash-can"></i> Hapus Riwayat
            </button>
        </div>
        
        <div id="sales-rating-header"></div>

        <!-- Daftar Pesanan -->'''
content = content.replace(header_old, header_new)

# 2. Add Logic to loadOrders
logic_old = '''        async function loadOrders() {
            if(!currentUid) return;
            const container = document.getElementById('orders-container');
            // Menampilkan loading state (skeleton loader)'''
            
logic_new = '''        async function loadOrders() {
            if(!currentUid) return;
            const container = document.getElementById('orders-container');
            
            // Calculate Sales Rating for Header
            try {
                const srRs = await turso.execute({
                    sql: "SELECT total_price, payment_method, status FROM riwayat_pemesanan WHERE seller_uid = ?",
                    args: [currentUid]
                });
                let revSum = 0;
                let succCount = 0;
                srRs.rows.forEach(r => {
                    const price = parseFloat(r.total_price) || 0;
                    if (r.status !== 'Retur Disetujui' && r.status !== 'Dibatalkan') {
                        succCount++;
                        if (r.payment_method === 'QRIS' || r.payment_method === 'PayPal') {
                            revSum += price;
                        }
                    }
                });
                let minSalesRating = 3.0 + Math.floor(succCount / 5) * 0.1 + Math.floor(revSum / 500000) * 0.1;
                if (minSalesRating > 5.0) minSalesRating = 5.0;
                
                const headerRatingEl = document.getElementById('sales-rating-header');
                if (headerRatingEl) {
                    headerRatingEl.innerHTML = `
                        <div class="bg-gradient-to-r from-amber-500 to-orange-500 p-4 rounded-2xl shadow-md text-white mb-6 flex justify-between items-center relative overflow-hidden">
                            <div class="absolute -right-4 -top-4 text-white/10 text-8xl">
                                <i class="fa-solid fa-star"></i>
                            </div>
                            <div class="z-10">
                                <p class="text-[10px] font-bold uppercase tracking-wider opacity-90 mb-0.5"><i class="fa-solid fa-wand-magic-sparkles"></i> Standar Rating Analisis AI</p>
                                <p class="text-xs opacity-90">Berdasarkan ${succCount} Pesanan Sukses & Rp ${new Intl.NumberFormat('id-ID').format(revSum)} Pendapatan</p>
                            </div>
                            <div class="flex flex-col items-center gap-0.5 bg-white/20 px-3 py-1.5 rounded-xl border border-white/30 backdrop-blur-sm z-10">
                                <div class="flex items-center gap-1.5">
                                    <i class="fa-solid fa-star text-yellow-300 text-lg"></i>
                                    <span class="font-display font-black text-2xl">${minSalesRating.toFixed(1)}</span>
                                </div>
                            </div>
                        </div>
                    `;
                }
            } catch(e) { console.error(e); }

            // Menampilkan loading state (skeleton loader)'''
content = content.replace(logic_old, logic_new)

with open(file_path, 'w', encoding='utf-8') as f:
    f.write(content)
print("Updated Riwayat Pemesanan with Sales Rating")
