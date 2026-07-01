# Gunakan image resmi PHP dengan server Apache
FROM php:8.1-apache

# Aktifkan modul rewrite (penting kalau web native kamu pakai URL routing)
RUN a2enmod rewrite

# Salin semua file dari GitHub ke dalam folder publik server
COPY . /var/www/html/

# Atur hak akses agar file bisa dibaca server
RUN chown -R www-data:www-data /var/www/html/

# Konfigurasi port agar sesuai dengan standar Google Cloud Run
RUN sed -i 's/80/${PORT}/g' /etc/apache2/sites-available/000-default.conf /etc/apache2/ports.conf
