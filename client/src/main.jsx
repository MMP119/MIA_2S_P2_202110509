import React from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter as Router, Route, Routes } from 'react-router-dom';
import App from './App.jsx';
import Login from './components/login/login.jsx';
import VisualDiscos from './components/discos/disks/discos.jsx';
import PartitionVisual from './components/discos/partitions/partitions.jsx'
import ArchivosVisualizador from './components/discos/archivos/archivos.jsx';
import TerminalUsuario from './components/terminalUsuario/terminalUsuario.jsx';
import './index.css';

ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <Router>
      <Routes>
        <Route path='/root' element={<App />} />
        <Route path='/' element={<Login />} />
        <Route path='/visualDiscos' element={<VisualDiscos />} />
        <Route path='/visualPartitions' element={<PartitionVisual />} />
        <Route path='/visualArchivos' element={<ArchivosVisualizador />} />
        <Route path='/terminalUsuario' element={<TerminalUsuario />} />
      </Routes>
    </Router>
  </React.StrictMode>,
)

