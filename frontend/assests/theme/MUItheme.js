import useMediaQuery from '@material-ui/core/useMediaQuery';
import { createMuiTheme, ThemeProvider } from '@material-ui/core/styles';
import { red } from '@material-ui/core/colors';

const theme = createMuiTheme({
    palette: {
        type: 'dark',
      primary: {
        main: '#e53935',
      },
      secondary: {
        main: '#1e88e5',
      },
      error: {
        main: red.A400,
      },  
    },
    overrides: {
      MuiTypography: {
        root: {
          color: '#ffffff',
        },
        colorTextPrimary: {
          color: '#cfcfcf',
        },
      },
      MuiToolbar: {
        root: {
          color: "#d32f2f"
        }
      }
      
    },
  });
  
  
  export default theme;