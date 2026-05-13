import { check, sleep } from 'k6';
import { SharedArray } from 'k6/data';
import encoding from 'k6/encoding';

export const options = {
    scenarios: {
        soak_test: {
            executor: 'ramping-vus',
            startVUs: 0,
            stages: [
                { duration: '2m', target: 500 },
                { duration: '5m', target: 2000 },
                { duration: '5m', target: 3500 },
                { duration: '5m', target: 5000 },
                { duration: '10m', target: 5000 },
                { duration: '2m', target: 3500 },
                { duration: '1m', target: 2000 },
                { duration: '1m', target: 500 },
                { duration: '30s', target: 0 },
            ],
            gracefulRampDown: '30s',
            gracefulStop: '30s',
        },
    },
    thresholds: {
        http_req_failed: ['rate<0.01'],
        http_req_duration: ['p(95)<5000', 'p(99)<10000'],
        iterations: ['count>100000'],
    },
    noConnectionReuse: false,
    discardResponseBodies: true,
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:20128';
const API_KEY = __ENV.API_KEY || 'test-key';

const models = new SharedArray('models', function() {
    return [
        'gpt-4o',
        'gpt-4o-mini',
        'claude-3-5-sonnet-20241022',
        'claude-3-5-haiku-20241022',
        'gemini-2.0-flash',
        'gemini-1.5-pro',
        'llama-3.1-70b-instruct',
        'qwen-2.5-72b-instruct',
    ];
});

export function setup() {
    console.log(`Starting soak test against ${BASE_URL}`);
    console.log('Target: 5000 concurrent SSE connections, 30 minutes');
    console.log('Memory target: <500MB RAM');
}

function randomModel() {
    return models[Math.floor(Math.random() * models.length)];
}

function jitter(min, max) {
    return min + Math.random() * (max - min);
}

function generateChatPayload(model) {
    return JSON.stringify({
        model: model,
        messages: [
            { role: 'system', content: 'You are a helpful assistant. Be concise.' },
            { role: 'user', content: `Say hello in exactly 3 words. Request ID: ${Date.now()}` },
        ],
        stream: true,
        max_tokens: 20,
    });
}

function generateNonStreamingPayload(model) {
    return JSON.stringify({
        model: model,
        messages: [
            { role: 'user', content: `What is 2+2? Just give the number. ID: ${Date.now()}` },
        ],
        stream: false,
        max_tokens: 5,
    });
}

export default function() {
    const model = randomModel();
    const isStreaming = Math.random() < 0.85;

    const params = {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${API_KEY}`,
        },
        timeout: '60s',
    };

    let response;
    let payload;

    if (isStreaming) {
        payload = generateChatPayload(model);
        params.headers['Accept'] = 'text/event-stream';
        response = http.post(`${BASE_URL}/v1/chat/completions`, payload, params);

        if (response.status === 200) {
            let eventCount = 0;
            const body = response.body;
            if (body) {
                const lines = body.split('\n');
                for (const line of lines) {
                    if (line.startsWith('data: ') && !line.includes('[DONE]')) {
                        eventCount++;
                    }
                }
            }
            check(response, {
                'streaming response received': (r) => r.status === 200,
                'has SSE events': () => eventCount > 0,
            });
        } else {
            check(response, {
                'streaming request succeeded': (r) => r.status === 200,
            });
        }
    } else {
        payload = generateNonStreamingPayload(model);
        response = http.post(`${BASE_URL}/v1/chat/completions`, payload, params);

        check(response, {
            'non-streaming request succeeded': (r) => r.status === 200,
        });
    }

    sleep(jitter(0.5, 2.0));
}

export function teardown(data) {
    console.log('Soak test completed');
}

export function handleSummary(data) {
    return {
        'stdout': JSON.stringify(data, null, 2),
        'load-test-results.json': JSON.stringify(data, null, 2),
    };
}
