import "./App.css";
import AppRoutes from "./AppRoutes";
import PreviewBanner from "./components/PreviewBanner";
import ThemeProvider from "./theme/ThemeProvider";

function App({ initialPalette }) {
  return (
    <ThemeProvider initialPalette={initialPalette}>
      <PreviewBanner />
      <AppRoutes />
    </ThemeProvider>
  );
}

export default App;
