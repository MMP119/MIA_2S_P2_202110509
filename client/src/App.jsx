// eslint-disable-next-line no-unused-vars
import React, { useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';

import './App.css';

function App() {
  const codeInputRef = useRef(null);
  const consoleOutputRef = useRef(null);
  const editorRef = useRef(null);
  const consoleEditorRef = useRef(null);
  const navigate = useNavigate();

  useEffect(() => {
    // Inicializa CodeMirror en el textarea con id 'codeInput'
    if (codeInputRef.current && !editorRef.current) {
      // eslint-disable-next-line no-undef
      editorRef.current = CodeMirror.fromTextArea(codeInputRef.current, {
        lineNumbers: true,
        mode: 'javascript',
        theme: 'dracula',
        viewportMargin: Infinity,
      });

      editorRef.current.getWrapperElement().style.fontSize = '18px';
    }

    // Inicializa CodeMirror en el textarea con id 'consoleOutput'
    if (consoleOutputRef.current && !consoleEditorRef.current) {
      // eslint-disable-next-line no-undef
      consoleEditorRef.current = CodeMirror.fromTextArea(consoleOutputRef.current, {
        lineNumbers: false,
        mode: 'text/plain',
        theme: 'dracula',
        readOnly: true,
        viewportMargin: Infinity,
      });
      consoleEditorRef.current.getWrapperElement().style.fontSize = '18px';
    }

    const openButton = document.getElementById('openButton');
    const runButton = document.getElementById('runButton');
    const clearButton = document.getElementById('clearButton');
    const closeButton = document.getElementById('closeSession');

    // función para el botón 'open'
    const openFile = () => {
      var input = document.createElement('input');
      input.type = 'file';
      input.onchange = e => {
        var file = e.target.files[0];
        var reader = new FileReader();
        reader.readAsText(file, 'UTF-8');
        reader.onload = readerEvent => {
          var content = readerEvent.target.result;
          editorRef.current.setValue(content);
        };
      };
      input.click();
    };

    // Función para el botón 'Run'
    const runCode = async () => {
      const code = editorRef.current.getValue();
      const commands = code.split('\n').filter(command => command.trim() !== '');
      let output = '';

      for (const command of commands) {
        if (command.toLowerCase().startsWith('rmdisk')) {
          const confirmation = window.confirm('¿Seguro que quiere eliminar el disco?');
          if (!confirmation) {
            output += `El comando RMDISK: ${command} fue cancelado por el usuario.\n`;
            continue; // Salta este comando si no se confirma
          }
        }

        try {
          const response = await fetch('http://54.80.109.226:8080/analyze', {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
            },
            body: JSON.stringify([command]), // Envía el comando individualmente
          });

          if (!response.ok) {
            const errorText = await response.text();
            output += `Error del servidor: ${errorText}\n`;
            continue; // Salta el procesamiento de este comando si hay error
          }

          const data = await response.json();

          // Agrega los resultados del comando al output
          if (data.results && typeof data.results === 'object') {
            output += `${Object.values(data.results).join('\n')}\n`;
          }

          // Agrega los errores del comando al output
          if (data.errors && typeof data.errors === 'object') {
            output += `${Object.values(data.errors).map(e => `Error - ${e}`).join('\n')}\n`;
          }
        } catch (error) {
          output += 'Error al conectar con el servidor.\n';
          console.error('Error:', error);
        }
      }

      consoleEditorRef.current.setValue(output || 'No hay salida');
    };


    // función para el botón 'Clear'
    const clearCode = () => {
      editorRef.current.setValue('');
      consoleEditorRef.current.setValue('');
    };

    const closeSession = () => {
      navigate('/');
    };

    openButton.addEventListener('click', openFile);
    runButton.addEventListener('click', runCode);
    clearButton.addEventListener('click', clearCode);
    closeButton.addEventListener('click', closeSession);

    // Cleanup event listeners on component unmount
    return () => {
      openButton.removeEventListener('click', openFile);
      runButton.removeEventListener('click', runCode);
      clearButton.removeEventListener('click', clearCode);
    };
  }, [navigate]);

  return (
    <div className="App">
      <div className="editor-container">
        <div className="header">
          <button id="openButton">
            <span className="material-symbols-outlined">upload</span>
          </button>
          <button id="clearButton">
            <span className="material-symbols-outlined">mop</span>
          </button>
          <button id="runButton">
            <span className="material-symbols-outlined">play_arrow</span>
          </button>
          <button id="closeSession">
            <span className="material-symbols-outlined">account_circle</span>
            Cerrar Sesión
          </button>
        </div>
        <div className="main">
          <div className="editor">
            <h3 id="textEditor">Code Input</h3>
            <textarea id="codeInput" ref={codeInputRef}></textarea>
          </div>
          <div className="console">
            <h3 id="textConsole">Console</h3>
            <textarea id="consoleOutput" ref={consoleOutputRef}></textarea>
          </div>
        </div>
      </div>
    </div>
  );
}

export default App;

