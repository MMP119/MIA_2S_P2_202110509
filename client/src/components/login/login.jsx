// eslint-disable-next-line no-unused-vars
import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import './login.css';

const Login = () => {
    const [partitionId, setPartitionId] = useState('');
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [login, setLogin] = useState({});
    const navigate = useNavigate();

    const handlePartitionIdChange = (e) => {
        setPartitionId(e.target.value);
    };

    const handleUsernameChange = (e) => {
        setUsername(e.target.value);
    };

    const handlePasswordChange = (e) => {
        setPassword(e.target.value);
    };


    const handleLogin = async() => {
        // implementar la lógica para autenticar al usuario
        if (partitionId === 'root' && username === 'root' && password === 'root') {
            console.log('Usuario autenticado');
            // redirigir al usuario a la página de comandos
            navigate('/root');
        }else if(partitionId === 'root' && username === 'root' && password === '123'){
            console.log('Usuario autenticado');
            // redirigir al usuario a la página de discos
            navigate('/visualDiscos?diskID=root');
        }

        try {
            const response = await fetch('http://localhost:8080/inicioSesion', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ partitionId, username, password }),
            });
            const data = await response.json();
            setLogin(data);

            console.log('Login:', partitionId, username, password);

            // Verifica si el login fue exitoso
            if (login === true) {
                console.log('Usuario autenticado');
                navigate(`/visualDiscos?diskID=${partitionId}`);
            } else {
                console.log('Credenciales incorrectas');
            }
        } catch (error) {
            console.error("Error fetching login:", error);
        }


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