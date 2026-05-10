import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    stages: [
        { duration: '10s', target: 500 },
        { duration: '30s', target: 1000 },
        { duration: '10s', target: 0 },
    ],
};

function generateUUID() {
    return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
        let r = Math.random() * 16 | 0, v = c == 'x' ? r : (r & 0x3 | 0x8);
        return v.toString(16);
    });
}

export default function () {
    // Aim at Kong on port 8000
    const url = 'http://localhost:8000/api/transfer?from=A&to=B&amount=1';

    const TOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJ1cGktaXNzdWVyIn0.p9gspgPRTr0AMC8RNAoaTKkGqFm_gKRwbP1S2uxZemk";

    const params = {
        headers: {
            'X-Idempotency-Key': generateUUID(),
            'Authorization': `Bearer ${TOKEN}`
        },
    };

    const res = http.post(url, null, params);

    check(res, {
        'is status 200': (r) => r.status === 200,
    });

    sleep(0.1); 
}