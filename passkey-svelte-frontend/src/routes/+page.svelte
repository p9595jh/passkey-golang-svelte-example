<style type="text/css">
    fieldset {
        margin: auto;
        margin-top: 5rem;
        padding: 1.5rem;
        width: 50%;
    }
    fieldset > input {
        margin-bottom: 0.5rem;
    }
</style>

<h1>Passkey Example</h1>

<div style="width: 100%;">
    <fieldset>
        <legend>Register</legend>
        <input bind:value={registerData.name} placeholder="name" /><br />
        <input bind:value={registerData.email} placeholder="email" /><br />
        <input bind:value={registerData.birthYear} placeholder="birth year" />
        <button on:click={() => register()}>submit</button>
    </fieldset>

    <fieldset>
        <legend>Login</legend>
        <input bind:value={loginData.name} placeholder="name" />
        <button on:click={() => login()}>submit</button>
    </fieldset>

    <fieldset>
        <legend>User Information</legend>
        <button on:click={() => requestUserData()}>request data</button>
        {#if userInfo}
            <br />
            <pre>{JSON.stringify(userInfo, null, 2)}</pre>
        {/if}
    </fieldset>
</div>

<script lang="ts">
    import { startRegistration, startAuthentication } from '@simplewebauthn/browser';

    const backend = 'http://localhost:4000';

    const registerData = {
        name: '',
        email: '',
        birthYear: '',
    };
    const loginData = {
        name: '',
    };
    let sessionId = '';
    let userInfo: any = undefined;

    async function register() {
        const { name, email, birthYear } = registerData;
        if (!name.length || !email.length || !birthYear.length) {
            alert('All fields need to be set!');
            return;
        }

        // call backend /register/start
        const res = await fetch(`${backend}/api/passkey/register/start`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                name,
                email,
                birthYear: Number(birthYear),
            }),
        });

        if (!res.ok) {
            const msg = await res.json();
            throw new Error(msg);
        }

        const data = await res.json();
        const attestationResponse = await startRegistration(
            data.options.publicKey
        ).catch((err) => {
            alert(String(err));
            return undefined;
        });
        if (!attestationResponse) return;

        const verificationResponse = await fetch(
            `${backend}/api/passkey/register/finish`,
            {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                    'X-Session-Id': data['sid'],
                },
                body: JSON.stringify(attestationResponse),
            },
        );

        const msg = await verificationResponse.json();
        alert(JSON.stringify(msg, null, 2));
    }

    async function login() {
        const { name } = loginData;
        if (!name.length) {
            alert('Name needs to be set!');
            return;
        }

        // call backend /login/start
        const res = await fetch(`${backend}/api/passkey/login/start`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name }),
        });

        if (!res.ok) {
            const msg = await res.json();
            throw new Error(msg);
        }

        const data = await res.json();
        const attestationResponse = await startAuthentication(
            data.options.publicKey
        ).catch((err) => {
            alert(String(err));
            return undefined;
        });
        if (!attestationResponse) return;

        const verificationResponse = await fetch(
            `${backend}/api/passkey/login/finish`,
            {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                    'X-Session-Id': data['sid'],
                },
                body: JSON.stringify(attestationResponse),
            },
        );

        const msg = await verificationResponse.json();
        alert(JSON.stringify(msg, null, 2));
        sessionId = msg.sid;
    }

    async function requestUserData() {
        const res = await fetch(`${backend}/api/forbidden`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-Session-Id': sessionId,
            },
        });
        userInfo = await res.json();
    }

</script>
