import { useContext, useState } from 'react';
import { Sidebar } from './components/Sidebar';
import { Topbar } from './components/Topbar';
import { ListFilms } from './pages/ListFilms';
import { ListSeries } from './pages/ListSeries';
import { MovieDetail } from './pages/MovieDetail';
import { SeriesDetail } from './pages/SeriesDetail';
import { UsersPage } from './pages/UsersPage';
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
  
  if (!context) return null;

  const { currentPage, selectedId, goToPage, goBack } = context;
  // const showBack = currentPage.startsWith('detail-');
  // const breadcrumb =
  //   currentPage === 'detail-movie'
  //     ? 'Films / Interstellar'
  //     : currentPage === 'detail-series'
  //       ? 'Séries / Breaking Bad'
  //       : '';

  return (
    <div className={layoutStyles.layout}>
      <Sidebar activeNav={currentPage} onNavClick={goToPage} />
      <div className={layoutStyles.main}>
        <Topbar 
          showBack={false} 
          breadcrumb="" 
          onBack={goBack}
          searchQuery={searchQuery}
          onSearchChange={setSearchQuery}
        />
        <div className={layoutStyles.content}>
          {currentPage === 'list-films' && (
            <div className={layoutStyles.page + ' ' + layoutStyles.active}>
              <ListFilms 
                onSelectMovie={(id) => goToPage('detail-movie', id)}
                searchQuery={searchQuery}
              />
            </div>
          )}
          {currentPage === 'list-series' && (
            <div className={layoutStyles.page + ' ' + layoutStyles.active}>
              <ListSeries 
                onSelectSeries={(id) => goToPage('detail-series', id)}
                searchQuery={searchQuery}
              />
            </div>
          )}
          {currentPage === 'detail-movie' && selectedId && (
            <div className={layoutStyles.page + ' ' + layoutStyles.active}>
              <MovieDetail movieId={selectedId} />
            </div>
          )}
          {currentPage === 'detail-series' && selectedId && (
            <div className={layoutStyles.page + ' ' + layoutStyles.active}>
              <SeriesDetail seriesId={selectedId} />
            </div>
          )}
          {currentPage === 'admin-users' && (
            <div className={layoutStyles.page + ' ' + layoutStyles.active}>
              <UsersPage />
            </div>
          )}
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
  if ((authMode === 'simple' || authMode === 'oidc') && !isAuthenticated) {
    return <LoginPage />;
  }

  // Show main app content
  return <AppContent />;
};

function App() {
  return (
    <AppContextProvider>
      <AppRouter />
    </AppContextProvider>
  );
}

export default App;
