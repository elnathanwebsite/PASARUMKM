/**
 * auth-google.js
 * ================================================================
 * Helper Autentikasi Google OAuth 2.0 (Google Identity Services)
 * Menggantikan Firebase Auth di semua halaman PASAR UMKM
 * 
 * Client ID: 327599206682-spog2c3u2566qqffdgjmu66dg1ja5109.apps.googleusercontent.com
 * ================================================================
 */

const GOOGLE_CLIENT_ID = '327599206682-spog2c3u2566qqffdgjmu66dg1ja5109.apps.googleusercontent.com';
const ADMIN_EMAIL      = 'arifanadmin@gmail.com';

// ── Storage Keys ────────────────────────────────────────────────
const KEY_USER    = 'pasar_umkm_user';    // { uid, email, name, photoURL, ... }
const KEY_TOKEN   = 'pasar_umkm_token';   // Google credential JWT
const KEY_ROLE    = 'pasar_umkm_role';    // 'user' | 'admin'

// ── Auth State Listeners ─────────────────────────────────────────
const _listeners = [];

/**
 * Daftarkan listener perubahan status auth
 * (mirip firebase.auth().onAuthStateChanged)
 * @param {function} cb  dipanggil dengan (user | null)
 */
function onAuthStateChanged(cb) {
    _listeners.push(cb);
    // Panggil segera dengan state saat ini
    const u = getCurrentUser();
    cb(u);
}

function _notifyListeners(user) {
    _listeners.forEach(cb => cb(user));
}

// ── Current User ─────────────────────────────────────────────────
/**
 * Mengembalikan objek user dari localStorage, atau null
 */
function getCurrentUser() {
    try {
        const raw = localStorage.getItem(KEY_USER);
        if (!raw) return null;
        const user = JSON.parse(raw);
        // Cek token masih ada
        if (!localStorage.getItem(KEY_TOKEN)) return null;
        return user;
    } catch { return null; }
}

// ── Sign Out ──────────────────────────────────────────────────────
function signOut() {
    localStorage.removeItem(KEY_USER);
    localStorage.removeItem(KEY_TOKEN);
    localStorage.removeItem(KEY_ROLE);
    try {
        sessionStorage.removeItem('has_shown_login_ad');
        Object.keys(sessionStorage).forEach(k => {
            if (k.startsWith('ad_shown_session_')) sessionStorage.removeItem(k);
        });
    } catch (_) {}
    // Revoke token Google (best-effort)
    try {
        google.accounts.oauth2.revoke(localStorage.getItem(KEY_TOKEN) || '', () => {});
    } catch (_) {}
    _notifyListeners(null);
}

// ── Decode JWT dari Google ────────────────────────────────────────
function _decodeJWT(token) {
    try {
        const base64Url = token.split('.')[1];
        const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
        const json = decodeURIComponent(atob(base64).split('').map(c =>
            '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2)
        ).join(''));
        return JSON.parse(json);
    } catch { return null; }
}

// ── Handle credential dari Google ────────────────────────────────
/**
 * Dipanggil setelah Google mengembalikan credential JWT
 * @param {string} credential  JWT dari Google
 * @param {string} role        'user' | 'admin'
 * @returns {{ user, isNewUser }}
 */
function handleGoogleCredential(credential, role = 'user') {
    const payload = _decodeJWT(credential);
    if (!payload) throw new Error('Token Google tidak valid');

    // Buat objek user (mirip Firebase user)
    const user = {
        uid       : payload.sub,              // Google subject = unique ID
        email     : payload.email,
        name      : payload.name,
        displayName: payload.name,
        photoURL   : payload.picture,
        emailVerified: payload.email_verified,
        role       : role
    };

    // Cek apakah user sudah pernah ada (isNewUser)
    const existingRaw = localStorage.getItem(KEY_USER);
    let isNewUser = true;
    if (existingRaw) {
        try {
            const existing = JSON.parse(existingRaw);
            if (existing.uid === user.uid) isNewUser = false;
        } catch (_) {}
    }

    // Simpan ke localStorage
    localStorage.setItem(KEY_USER, JSON.stringify(user));
    localStorage.setItem(KEY_TOKEN, credential);
    localStorage.setItem(KEY_ROLE, role);

    _notifyListeners(user);
    return { user, isNewUser };
}

// ── Inisialisasi Google Identity Services ─────────────────────────
/**
 * Init GIS dan render tombol Google (jika container ada)
 * @param {object} options
 * @param {string} options.buttonId      ID elemen container tombol Google
 * @param {function} options.onSuccess   Dipanggil dengan ({ user, isNewUser })
 * @param {function} options.onError     Dipanggil dengan error
 * @param {string} [options.role]        'user' | 'admin'
 * @param {string} [options.context]     'signin' | 'signup' | 'use'
 */
function initGoogleSignIn({ buttonId, onSuccess, onError, role = 'user', context = 'signin' }) {
    if (typeof google === 'undefined') {
        console.error('[auth-google] Google Identity Services tidak dimuat');
        return;
    }

    google.accounts.id.initialize({
        client_id       : GOOGLE_CLIENT_ID,
        callback        : (response) => {
            try {
                const result = handleGoogleCredential(response.credential, role);
                onSuccess && onSuccess(result);
            } catch (err) {
                onError && onError(err);
            }
        },
        auto_select     : false,
        cancel_on_tap_outside: true,
    });

    // Render tombol jika ada container
    const container = document.getElementById(buttonId);
    if (container) {
        google.accounts.id.renderButton(container, {
            type     : 'standard',
            theme    : 'outline',
            size     : 'large',
            text     : context === 'signup' ? 'signup_with' : 'signin_with',
            shape    : 'rectangular',
            logo_alignment: 'left',
            width    : container.offsetWidth || 300,
        });
    }
}

/**
 * Tampilkan One Tap (popup otomatis Google)
 */
function showOneTap(onSuccess, onError, role = 'user') {
    if (typeof google === 'undefined') return;
    google.accounts.id.initialize({
        client_id : GOOGLE_CLIENT_ID,
        callback  : (response) => {
            try {
                const result = handleGoogleCredential(response.credential, role);
                onSuccess && onSuccess(result);
            } catch (err) {
                onError && onError(err);
            }
        },
        auto_select: false,
    });
    google.accounts.id.prompt((notification) => {
        if (notification.isNotDisplayed()) {
            console.log('[auth-google] One Tap tidak ditampilkan:', notification.getNotDisplayedReason());
        }
        if (notification.isDismissedMoment()) {
            console.log('[auth-google] One Tap ditutup:', notification.getDismissedReason());
        }
    });
}

/**
 * Trigger Google Sign-In popup secara manual (untuk tombol kustom)
 * Menggunakan OAuth2 Implicit Flow
 * @param {function} onSuccess   ({ user, isNewUser })
 * @param {function} onError     (err)
 * @param {string}   role        'user' | 'admin'
 */
function signInWithGoogle(onSuccess, onError, role = 'user') {
    if (typeof google === 'undefined') {
        onError && onError(new Error('Google Identity Services belum dimuat'));
        return;
    }

    const client = google.accounts.oauth2.initTokenClient({
        client_id: GOOGLE_CLIENT_ID,
        scope    : 'openid profile email',
        callback : (tokenResponse) => {
            if (tokenResponse.error) {
                onError && onError(new Error(tokenResponse.error));
                return;
            }
            // Ambil info user dari userinfo endpoint
            fetch('https://www.googleapis.com/oauth2/v3/userinfo', {
                headers: { Authorization: `Bearer ${tokenResponse.access_token}` }
            })
            .then(r => r.json())
            .then(payload => {
                const user = {
                    uid          : payload.sub,
                    email        : payload.email,
                    name         : payload.name,
                    displayName  : payload.name,
                    photoURL     : payload.picture,
                    emailVerified: payload.email_verified,
                    role         : role
                };
                const existingRaw = localStorage.getItem(KEY_USER);
                let isNewUser = true;
                if (existingRaw) {
                    try {
                        const existing = JSON.parse(existingRaw);
                        if (existing.uid === user.uid) isNewUser = false;
                    } catch (_) {}
                }
                localStorage.setItem(KEY_USER, JSON.stringify(user));
                localStorage.setItem(KEY_TOKEN, tokenResponse.access_token);
                localStorage.setItem(KEY_ROLE, role);
                _notifyListeners(user);
                onSuccess && onSuccess({ user, isNewUser });
            })
            .catch(err => onError && onError(err));
        },
    });
    client.requestAccessToken({ prompt: 'select_account' });
}

// ── Export ke global window ───────────────────────────────────────
window.PasarAuth = {
    GOOGLE_CLIENT_ID,
    ADMIN_EMAIL,
    onAuthStateChanged,
    getCurrentUser,
    signOut,
    signInWithGoogle,
    initGoogleSignIn,
    showOneTap,
    handleGoogleCredential,
};
