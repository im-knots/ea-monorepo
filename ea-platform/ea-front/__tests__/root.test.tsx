import '@testing-library/jest-dom';
import { render, screen } from '@testing-library/react';
import RootPage from '../app/page';
import LoginPage from '../app/login/page';

describe('Page', () => {
  it('renders the root page', () => {
    render(<RootPage />)
  })

  it('renders the login page', () => {
    render(<LoginPage />)
  })
})