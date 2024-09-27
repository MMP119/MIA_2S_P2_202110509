// eslint-disable-next-line no-unused-vars
import React, { useState, useEffect } from 'react';
// eslint-disable-next-line no-unused-vars
import { useNavigate } from 'react-router-dom'; // Para redireccionar a la vista de particiones
import './archivos.css';


const ArchivosVisualizador = () => {

    const[archivo, setArchivos] = useState([]); // Almacena los archivos de la carpeta
    const[loading, setLoading] = useState(true); // Para manejar el estado de carga
    //const navigate = useNavigate(); // Hook para redireccionar
    // eslint-disable-next-line no-unused-vars
    const [ruta, setRuta] = useState("");

    // ontener el id de la particion desde la URL y el path
    const query = new URLSearchParams(location.search);
    const idParticion = query.get('partitionId');
    const path = ruta;

    // Función para obtener los archivos de la carpeta desde el backend
    useEffect(() => {
        const fetchArchivos = async () => {
            try {
                const response = await fetch('http://localhost:8080/archivosCarpetas', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({ idParticion, path }),  // Enviar el id de la particion y el path al backend
                    
                });
                console.log(path)
                const data = await response.json();
                setArchivos(data);  // Actualiza el estado con los archivos obtenidos
                setLoading(false);
            } catch (error) {
                console.error("Error fetching archivos:", error);
                setLoading(false);
            }
        };

        fetchArchivos();  // Llama a la función al montar el componente
    }, [idParticion, path]);


    // Mostrar un mensaje de carga mientras se obtienen los archivos
    if (loading) {
        return <div className='discos'><h1>Cargando archivos...</h1></div>;
    }

    const handleFileClick = (fileName, isFolder) => {
        if(isFolder){
            let newPath = `${ruta}/${fileName}`;
            setRuta(newPath);
            //limpiar la ruta
            newPath = "";
        }
    };

    
    return (
        <div className='discos'>
            <div className='discos-container'>
                <h1>Visualizador del Sistema de Archivos</h1>
                <p>Navegue entre carpetas o visualice archivos</p>
                <br></br>
                <textarea id="ruta"  rows="2" cols="100" readOnly value={ruta}></textarea>
                <br></br>
                <br></br>
                <div className="discos-grid">
                    {Array.isArray(archivo) && archivo.length > 0 &&(
                        archivo.map((file, index) =>(
                            <button 
                                key={index} 
                                className='disco'
                                onClick={() => handleFileClick(file.nombre, file.tipo)}  // Manejador de clic para el archivo
                            >
                                <span className="material-symbols-outlined">
                                    {file.tipo =='archivo' ? 'insert_drive_file' : 'folder'}
                                </span>
                                {file.nombre}
                            </button>
                        ))
                        )
                    }
                </div>
            </div>
        </div>
    );

}

export default ArchivosVisualizador;