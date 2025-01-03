// eslint-disable-next-line no-unused-vars
import React, { useState, useEffect } from 'react';
import { useLocation } from 'react-router-dom'; // Para obtener el parámetro de la URL
import { useNavigate } from 'react-router-dom';
import './partitions.css';

const ParticionVisualizador = () => {
    const [particiones, setParticiones] = useState([]);  // Almacena las particiones del disco
    const [loading, setLoading] = useState(true);        // Maneja el estado de carga
    const location = useLocation();
    const navigate = useNavigate();  // Hook para redireccionar

    // Obtener el nombre del disco desde la URL
    const query = new URLSearchParams(location.search);
    const diskName = query.get('diskName');

    // Función para obtener las particiones desde el backend
    useEffect(() => {
        const fetchParticiones = async () => {
            try {
                const response = await fetch('http://54.80.109.226:8080/partitions', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({ diskName }),  // Enviar el nombre del disco al backend
                });
                const data = await response.json();
                setParticiones(data);  // Actualiza el estado con las particiones obtenidas
                setLoading(false);
            } catch (error) {
                console.error("Error fetching partitions:", error);
                setLoading(false);
            }
        };

        fetchParticiones();  // Llama a la función al montar el componente
    }, [diskName]);

    // Mostrar un mensaje de carga mientras se obtienen las particiones
    if (loading) {
        return <div className='discos'><h1>Cargando particiones...</h1></div>;
    }

    //Funcion que maneja el clic de las particiones
    const handleClick = async (particionId) => {
        try {
            // Redirigir a la vista de archivos con el nombre de la partición seleccionada
            navigate(`/visualArchivos?partitionId=${particionId}`);
        } catch (error) {
            console.error("Error al redirigir a la vista de archivos:", error);
            setLoading(false);
        }
    };

    return (
        <div className='discos'>
            <button className='back-button' onClick={() => navigate('/')}><span className="material-symbols-outlined">arrow_back</span> Volver</button>
            <div className='discos-container'>
                <h1>Particiones del Disco {diskName}</h1>
                <p>Seleccione la partición que desea visualizar:</p>
                <br></br>
                <br></br>
                <div className="discos-grid">
                    {Array.isArray(particiones) && particiones.length > 0 ? (
                        particiones.map((particion, index) => (
                            <button 
                                key={index} 
                                className='disco'
                                onClick={() => handleClick(particion.partitionId, particion.path)}  // Manejador de clic para la partición
                            >
                                <span className="material-symbols-outlined">clock_loader_40</span>
                                {particion.partitionName}
                            </button>
                        ))
                    ) : (
                        <p>No se encontraron particiones para este disco.</p>
                    )}
                </div>
            </div>
        </div>
    );
};

export default ParticionVisualizador;