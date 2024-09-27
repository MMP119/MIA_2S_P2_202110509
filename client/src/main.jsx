import React from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter as Router, Route, Routes } from 'react-router-dom';
import App from './App.jsx';
import Login from './components/login/login.jsx';
import VisualDiscos from './components/discos/disks/discos.jsx';
import PartitionVisual from './components/discos/partitions/partitions.jsx'
import './index.css';

ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <Router>
      <Routes>
        <Route path='/' element={<App />} />
        <Route path='/login' element={<Login />} />
        <Route path='/visualDiscos' element={<VisualDiscos />} />
        <Route path='/visualPartitions' element={<PartitionVisual />} />
      </Routes>
    </Router>
  </React.StrictMode>,
)

