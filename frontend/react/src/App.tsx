import { useContext, useMemo, useState } from 'react';
import { BrowserRouter, Navigate, Route, Routes, useLocation, useNavigate, useParams } from 'react-router-dom';
import { Sidebar } from './components/Sidebar';
import { Topbar } from './components/Topbar';
import { ListFilms } from './pages/ListFilms';
import { ListSeries } from './pages/ListSeries';
import { MovieDetail } from './pages/MovieDetail';
import { SeriesDetail } from './pages/SeriesDetail';
import { LoginPage } from './pages/LoginPage';
import { AppContext, AppContextProvider } from './hooks/useAppContext.tsx';
import layoutStyles from './styles/layout.module.css';
import './styles/variables.css';

// Loading spinner for auth check
const AuthLoadingSpinner = () => (
  <div style={{
    minHeight: '100vh',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    background: 'var(--color-background-tertiary)',
  }}>
    <div style={{
      width: '32px',
      height: '32px',
      border: '3px solid var(--color-border-tertiary)',
      borderTopColor: 'var(--color-primary)',
      borderRadius: '50%',
      animation: 'spin 0.8s linear infinite',
    }} />
    <style>{`
      @keyframes spin {
        to { transform: rotate(360deg); }
      }
    `}</style>
  </div>
);

const AppContent = () => {
  const context = useContext(AppContext);
  const [searchQuery, setSearchQuery] = useState('');
  const location = useLocation();
  const navigate = useNavigate();
  
  if (!context) return null;

  const activeNav = useMemo<'movies' | 'series'>(() => {
    return location.pathname.startsWith('/series') ? 'series' : 'movies';
  }, [location.pathname]);

  const handleSidebarNav = (page: 'movies' | 'series') => {
    navigate(page === 'movies' ? '/movies' : '/series');
  };

  const handleBack = () => {
    if (window.history.length > 1) {
      navigate(-1);
      return;
    }

    navigate(activeNav === 'series' ? '/series' : '/movies');
  };

  return (
    <div className={layoutStyles.layout}>
      <Sidebar activeNav={activeNav} onNavClick={handleSidebarNav} />
      <div className={layoutStyles.main}>
        <Topbar 
          showBack={false} 
          breadcrumb="" 
          onBack={handleBack}
          searchQuery={searchQuery}
          onSearchChange={setSearchQuery}
        />
        <div className={layoutStyles.content}>
          <Routes>
            <Route
              path="/"
              element={<Navigate to="/movies" replace />}
            />
            <Route
              path="/movies"
              element={(
                <div className={layoutStyles.page + ' ' + layoutStyles.active}>
                  <ListFilms
                    onSelectMovie={(id) => navigate(`/movies/${id}`)}
                    searchQuery={searchQuery}
                  />
                </div>
              )}
            />
            <Route
              path="/movies/:movieId"
              element={(
                <div className={layoutStyles.page + ' ' + layoutStyles.active}>
                  <MovieDetailFromRoute />
                </div>
              )}
            />
            <Route
              path="/series"
              element={(
                <div className={layoutStyles.page + ' ' + layoutStyles.active}>
                  <ListSeries
                    onSelectSeries={(id) => navigate(`/series/${id}`)}
                    searchQuery={searchQuery}
                  />
                </div>
              )}
            />
            <Route
              path="/series/:seriesId"
              element={(
                <div className={layoutStyles.page + ' ' + layoutStyles.active}>
                  <SeriesDetailFromRoute />
                </div>
              )}
            />
            <Route
              path="*"
              element={<Navigate to="/movies" replace />}
            />
          </Routes>
        </div>
      </div>
    </div>
  );
};

// Auth-aware router
const AppRouter = () => {
  const context = useContext(AppContext);
  
  if (!context) return null;

  const { authLoading, isAuthenticated, authMode } = context;

  // Show loading spinner while checking auth
  if (authLoading) {
    return <AuthLoadingSpinner />;
  }

  // Show login page if auth required and not authenticated
  if (authMode === 'simple' && !isAuthenticated) {
    return <LoginPage />;
  }

  // Show main app content
  return <AppContent />;
};

const MovieDetailFromRoute = () => {
  const { movieId } = useParams();
  const parsedId = Number(movieId);

  if (!movieId || Number.isNaN(parsedId)) {
    return <div style={{ padding: '20px' }}>Film non trouvé</div>;
  }

  return <MovieDetail movieId={parsedId} />;
};

const SeriesDetailFromRoute = () => {
  const { seriesId } = useParams();
  const parsedId = Number(seriesId);

  if (!seriesId || Number.isNaN(parsedId)) {
    return <div style={{ padding: '20px' }}>Série non trouvée</div>;
  }

  return <SeriesDetail seriesId={parsedId} />;
};

function App() {
  return (
    <BrowserRouter>
      <AppContextProvider>
        <AppRouter />
      </AppContextProvider>
    </BrowserRouter>
  );
}

export default App;
