import { Container, Typography, Paper, Table, TableBody, TableCell, TableContainer, TableHead, TableRow } from '@mui/material'
import { useQuery } from '@tanstack/react-query'
import axios from 'axios'

interface LeaderboardEntry {
  rank: number
  username: string
  points: number
  correctPredictions: number
  totalPredictions: number
}

const Leaderboard = () => {
  const { data: leaderboard, isLoading } = useQuery<LeaderboardEntry[]>({
    queryKey: ['leaderboard'],
    queryFn: async () => {
      const response = await axios.get('/api/leaderboard')
      return response.data
    },
  })

  return (
    <Container maxWidth="lg">
      <Typography variant="h4" component="h1" gutterBottom>
        Leaderboard
      </Typography>

      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>Rank</TableCell>
              <TableCell>Username</TableCell>
              <TableCell align="right">Points</TableCell>
              <TableCell align="right">Correct Predictions</TableCell>
              <TableCell align="right">Success Rate</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {isLoading ? (
              <TableRow>
                <TableCell colSpan={5} align="center">
                  Loading leaderboard...
                </TableCell>
              </TableRow>
            ) : (
              leaderboard?.map((entry) => (
                <TableRow key={entry.username}>
                  <TableCell>{entry.rank}</TableCell>
                  <TableCell>{entry.username}</TableCell>
                  <TableCell align="right">{entry.points}</TableCell>
                  <TableCell align="right">{entry.correctPredictions}</TableCell>
                  <TableCell align="right">
                    {((entry.correctPredictions / entry.totalPredictions) * 100).toFixed(1)}%
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </TableContainer>
    </Container>
  )
}

export default Leaderboard
