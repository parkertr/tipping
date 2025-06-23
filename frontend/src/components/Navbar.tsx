import { AppBar, Toolbar, Typography, Button, Box } from '@mui/material'
import { Link as RouterLink } from 'react-router-dom'
import SportsSoccerIcon from '@mui/icons-material/SportsSoccer'
import GoogleSignIn from './GoogleSignIn'
import UserMenu from './UserMenu'
import { useAuth } from '../contexts/AuthContext'

const Navbar = () => {
  const { isAuthenticated, isLoading } = useAuth()

  return (
    <AppBar position="static">
      <Toolbar>
        <SportsSoccerIcon sx={{ mr: 2 }} />
        <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
          Footy Tipping
        </Typography>

        {/* Navigation Links - only show for authenticated users */}
        {isAuthenticated && (
          <Box sx={{ display: 'flex', mr: 2 }}>
            <Button color="inherit" component={RouterLink} to="/">
              Home
            </Button>
            <Button color="inherit" component={RouterLink} to="/matches">
              Matches
            </Button>
            <Button color="inherit" component={RouterLink} to="/leaderboard">
              Leaderboard
            </Button>
            <Button color="inherit" component={RouterLink} to="/profile">
              Profile
            </Button>
          </Box>
        )}

        {/* Authentication UI */}
        <Box>
          {!isLoading && (
            isAuthenticated ? (
              <UserMenu />
            ) : (
              <GoogleSignIn />
            )
          )}
        </Box>
      </Toolbar>
    </AppBar>
  )
}

export default Navbar
