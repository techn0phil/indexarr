import { useContext, useState } from 'react';
import { Sidebar } from './components/Sidebar';
import { Topbar } from './components/Topbar';
import { ListFilms } from './pages/ListFilms';
import { ListSeries } from './pages/ListSeries';
import { Recents } from './pages/Recents';
import { Statistics } from './pages/Statistics';
import { Problems } from './pages/Problems';
import { MovieDetail } from './pages/MovieDetail';
import { SeriesDetail } from './pages/SeriesDetail';
import { AppContext, AppContextProvider } from './hooks/useAppContext.tsx';
import layoutStyles from './styles/layout.module.css';
import './styles/variables.css';

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
          {currentPage === 'list-recents' && (
            <div className={layoutStyles.page + ' ' + layoutStyles.active}>
              <Recents
                onSelectMovie={(id) => goToPage('detail-movie', id)}
                onSelectSeries={(id) => goToPage('detail-series', id)}
              />
            </div>
          )}
          {currentPage === 'statistics' && (
            <div className={layoutStyles.page + ' ' + layoutStyles.active}>
              <Statistics />
            </div>
          )}
          {currentPage === 'problems' && (
            <div className={layoutStyles.page + ' ' + layoutStyles.active}>
              <Problems />
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
        </div>
      </div>
    </div>
  );
};

function App() {
  return (
    <AppContextProvider>
      <AppContent />
    </AppContextProvider>
  );
}

export default App;
