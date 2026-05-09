import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    // Stage 1: Fast ramp up to 500 users
    // Stage 2: Push to 1000 users and hold it there
    // Stage 3: Drop back down
    stages: [
        { duration: '10s', target: 500 },
        { duration: '30s', target: 1000 },
        { duration: '10s', target: 0 },
    ],
};

// Pure JS UUID generator to avoid external HTTP imports failing
function generateUUID() {
    return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
        let r = Math.random() * 16 | 0, v = c == 'x' ? r : (r & 0x3 | 0x8);
        return v.toString(16);
    });
}

// THIS is the main execution loop k6 is looking for
export default function () {
    const url = 'http://localhost:8081/api/transfer?from=A&to=B&amount=1';
    
    const params = {
        headers: {
            'X-Idempotency-Key': generateUUID(),
        },
    };

    // Fire the POST request (null is the body since we use query params)
    const res = http.post(url, null, params);

    // Check if the transaction was successful
    check(res, {
        'is status 200': (r) => r.status === 200,
    });

    // A tiny sleep to simulate real-world user pacing
    sleep(0.1); 
}