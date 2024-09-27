// eslint-disable-next-line no-unused-vars
import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom'; // Para redireccionar a la vista de particiones
import './discos.css';

const DiscosVisualizador = () => {
    const [discos, setDiscos] = useState({})    // Almacena los discos obtenidos del servidor
    const [loading, setLoading] = useState(true) // Para manejar el estado de carga
    const navigate = useNavigate();  // Hook para redireccionar

    // Función para obtener los discos desde el backend
    useEffect(() => {
        const fetchDiscos = async () => {
            try {
                const response = await fetch('http://localhost:8080/disks');  // Endpoint del backend
                const data = await response.json();
                setDiscos(data);  // Actualiza el estado con los discos obtenidos
                setLoading(false);  // Termina la carga
            } catch (error) {
                console.error("Error fetching disks:", error);
                setLoading(false);  // Termina la carga en caso de error
            }
        };

        fetchDiscos();  // Llama a la función al montar el componente
    }, []);

    // Función que maneja el clic de los discos
    const handleClick = async (disco) => {
        try {
            // Redirigir a la vista de particiones con el nombre del disco seleccionado
            navigate(`/visualPartitions?diskName=${disco}`);
        } catch (error) {
            console.error("Error al redirigir a la vista de particiones:", error);
        }
    };

    // Mostrar un mensaje de carga mientras se obtienen los discos
    if (loading) {
        return <div className='discos'><h1>Cargando discos...</h1></div>;
    }

    return (
        <div className='discos'>
            <div className='discos-container'>
                <h1>Visualizador del Sistema de Archivos</h1>
                <p>Seleccione el disco que desea visualizar:</p>
                <br></br>
                <br></br>
                <div className="discos-grid">
                    {Object.keys(discos).map((disco, index) => (
                        <button 
                            key={index} 
                            className='disco'
                            onClick={() => handleClick(disco)}  // Llama a la función con el disco seleccionado
                        >
                        <span className="material-symbols-outlined">hard_drive</span>
                        {disco}
                        </button>
                    ))}
                </div>
            </div>
        </div>
    );
};

export default DiscosVisualizador;