// eslint-disable-next-line no-unused-vars
import React, { useState } from 'react';
import './login.css';

const Login = () => {
    const [partitionId, setPartitionId] = useState('');
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');

    const handlePartitionIdChange = (e) => {
        setPartitionId(e.target.value);
    };

    const handleUsernameChange = (e) => {
        setUsername(e.target.value);
    };

    const handlePasswordChange = (e) => {
        setPassword(e.target.value);
    };

    const handleLogin = () => {
        // Aquí puedes implementar la lógica para autenticar al usuario
        console.log('Partition ID:', partitionId);
        console.log('Username:', username);
        console.log('Password:', password);
    };

    return (
        <div className='login'>
            <div className='login-container'>
                <h2>Login</h2>
                <div className="input-group">
                    <label htmlFor="partitionId">ID Partición:</label>
                    <input type="text" id="partitionId" value={partitionId} onChange={handlePartitionIdChange} />
                </div>
                <div className="input-group">
                    <label htmlFor="username">Usuario:</label>
                    <input type="text" id="username" value={username} onChange={handleUsernameChange} />
                </div>
                <div className="input-group">
                    <label htmlFor="password">Contraseña:</label>
                    <input type="password" id="password" value={password} onChange={handlePasswordChange} />
                </div>
                <button onClick={handleLogin}>Iniciar sesión</button>
            </div>
        </div>
    );
};

export default Login;