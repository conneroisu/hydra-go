/**
 * Stop Hook
 * 
 * Comprehensive session completion handler with quality assurance operations.
 * Executes linting, testing, and formatting in parallel for optimal performance.
 * Provides detailed error reporting and maintains strict type safety.
 * 
 * @fileoverview Stop hook implementation with full TypeScript typing and error handling
 * @author Claude Code Hook System
 * @version 1.0.0
 */

import type { HookResult } from '../types.ts';
import {
  createHookResult,
  handleError,
  executeShellCommand,
  Logger,
  PerformanceMonitor,
} from '../utils.ts';

// ============================================================================
// Types and Interfaces
// ============================================================================

/**
 * Result of a quality assurance operation (tests, linting, formatting)
 */
interface QualityAssuranceResult {
  readonly success: boolean;
  readonly errorMessage?: string;
  readonly duration?: number;
  readonly exitCode?: number;
}

/**
 * Configuration for quality assurance operations
 */
interface QualityAssuranceConfig {
  readonly timeout: number;
  readonly command: string;
  readonly errorLogFile: string;
}

/**
 * Execution context for quality assurance operations
 */
interface ExecutionContext {
  readonly operationType: QualityOperationType;
  readonly startTime: number;
}

// ============================================================================
// Constants
// ============================================================================

/**
 * Quality assurance operation types
 */
type QualityOperationType = 'linting' | 'testing' | 'formatting';

/**
 * Default timeout values for different operations (in milliseconds)
 */
const OPERATION_TIMEOUTS = {
  LINTING: 120_000,
  TESTING: 120_000,
  FORMATTING: 60_000,
} as const;

/**
 * Detects the appropriate test command based on project type
 */
function detectTestCommand(): string {
  try {
    const fs = require('fs');
    
    // Check if go.mod exists (Go project)
    if (fs.existsSync('go.mod')) {
      return 'go test -v ./...';
    }
    
    // Check if package.json exists (Node.js/TypeScript project)
    if (fs.existsSync('package.json')) {
      return 'bun test';
    }
    
    // Check if Cargo.toml exists (Rust project)
    if (fs.existsSync('Cargo.toml')) {
      return 'cargo test';
    }
    
    // Default to bun test for TypeScript projects
    return 'bun test';
  } catch {
    // Fallback to bun test if detection fails
    return 'bun test';
  }
}

/**
 * Shell commands for quality assurance operations
 */
const QA_COMMANDS = {
  LINTING: 'nix develop -c lint',
  TESTING: detectTestCommand(),
  FORMATTING: 'nix fmt',
} as const;

/**
 * Error log file names for different operations
 */
const ERROR_LOG_FILES = {
  LINTING: 'linting-error.txt',
  TESTING: 'tests-error.txt',
  FORMATTING: 'formatting-error.txt',
} as const;

/**
 * Success and error messages
 */
const MESSAGES = {
  SUCCESS: 'Session completed successfully',
  LINTING_FAILED: 'Stop hook failed: Linting did not pass',
  TESTING_FAILED: 'Stop hook failed: Tests did not pass',
  FORMATTING_FAILED: 'Stop hook failed: Formatting did not pass',
  NO_OUTPUT: 'No output available',
} as const;

// ============================================================================
// Error Classes
// ============================================================================

/**
 * Base class for quality assurance errors
 */
abstract class QualityAssuranceError extends Error {
  public readonly operationType: QualityOperationType;
  public readonly exitCode?: number;
  public readonly duration?: number;

  constructor(
    message: string,
    operationType: QualityOperationType,
    exitCode?: number,
    duration?: number
  ) {
    super(message);
    this.name = this.constructor.name;
    this.operationType = operationType;
    this.exitCode = exitCode;
    this.duration = duration;
    
    // Ensure proper prototype chain for instanceof checks
    Object.setPrototypeOf(this, new.target.prototype);
  }
}

/**
 * Error thrown when linting operation fails
 */
class LintingError extends QualityAssuranceError {
  constructor(message: string, exitCode?: number, duration?: number) {
    super(message, 'linting', exitCode, duration);
  }
}

/**
 * Error thrown when testing operation fails
 */
class TestingError extends QualityAssuranceError {
  constructor(message: string, exitCode?: number, duration?: number) {
    super(message, 'testing', exitCode, duration);
  }
}

/**
 * Error thrown when formatting operation fails
 */
class FormattingError extends QualityAssuranceError {
  constructor(message: string, exitCode?: number, duration?: number) {
    super(message, 'formatting', exitCode, duration);
  }
}

// ============================================================================
// Stop Hook Implementation
// ============================================================================

/**
 * Stop hook class responsible for session completion and quality assurance
 * 
 * This class orchestrates the execution of linting, testing, and formatting
 * operations in parallel to ensure code quality before session completion.
 * 
 * @example
 * ```typescript
 * const result = await StopHook.execute();
 * if (!result.success) {
 *   console.error(result.message);
 *   process.exit(result.exit_code);
 * }
 * ```
 */
export class StopHook {
  /**
   * Executes the stop hook with comprehensive quality assurance
   * 
   * @returns Promise resolving to HookResult with execution status
   * @throws Never throws - all errors are caught and returned as failed HookResult
   */
  static async execute(): Promise<HookResult> {
    const startTime = PerformanceMonitor.startTiming('stop');
    
    try {
      Logger.info('Starting stop hook execution');
      
      // Execute quality assurance operations in parallel for optimal performance
      const [lintingResult, testsResult] = await Promise.allSettled([
        this.executeLinting(),
        this.executeTesting(),
      ]);
      
      // Process linting results
      if (lintingResult.status === 'rejected' || !lintingResult.value.success) {
        const error = lintingResult.status === 'rejected' 
          ? lintingResult.reason 
          : lintingResult.value.errorMessage;
        const errorMessage = this.formatErrorMessage(MESSAGES.LINTING_FAILED, error);
        
        PerformanceMonitor.endTiming('stop', startTime, false, errorMessage);
        return createHookResult(false, errorMessage, true);
      }
      
      // Process testing results
      if (testsResult.status === 'rejected' || !testsResult.value.success) {
        const error = testsResult.status === 'rejected' 
          ? testsResult.reason 
          : testsResult.value.errorMessage;
        const errorMessage = this.formatErrorMessage(MESSAGES.TESTING_FAILED, error);
        
        PerformanceMonitor.endTiming('stop', startTime, false, errorMessage);
        return createHookResult(false, errorMessage, true);
      }
      
      // Execute formatting as final step
      const formattingResult = await this.executeFormatting();
      if (!formattingResult.success) {
        const errorMessage = this.formatErrorMessage(
          MESSAGES.FORMATTING_FAILED, 
          formattingResult.errorMessage
        );
        
        PerformanceMonitor.endTiming('stop', startTime, false, errorMessage);
        return createHookResult(false, errorMessage, true);
      }
      
      Logger.info('Stop hook completed successfully');
      PerformanceMonitor.endTiming('stop', startTime, true);
      return createHookResult(true, MESSAGES.SUCCESS);
      
    } catch (error) {
      const errorMessage = this.extractErrorMessage(error);
      PerformanceMonitor.endTiming('stop', startTime, false, errorMessage);
      return handleError(error, 'stop hook');
    }
  }
  
  // ==========================================================================
  // Quality Assurance Operations
  // ==========================================================================
  
  /**
   * Executes testing operation with comprehensive error handling
   * 
   * @returns Promise resolving to QualityAssuranceResult
   * @private
   */
  private static async executeTesting(): Promise<QualityAssuranceResult> {
    const config: QualityAssuranceConfig = {
      timeout: OPERATION_TIMEOUTS.TESTING,
      command: QA_COMMANDS.TESTING,
      errorLogFile: ERROR_LOG_FILES.TESTING,
    };
    
    return this.executeQualityAssuranceOperation('testing', config);
  }
  
  /**
   * Executes linting operation with comprehensive error handling
   * 
   * @returns Promise resolving to QualityAssuranceResult
   * @private
   */
  private static async executeLinting(): Promise<QualityAssuranceResult> {
    const config: QualityAssuranceConfig = {
      timeout: OPERATION_TIMEOUTS.LINTING,
      command: QA_COMMANDS.LINTING,
      errorLogFile: ERROR_LOG_FILES.LINTING,
    };
    
    return this.executeQualityAssuranceOperation('linting', config);
  }
  
  /**
   * Executes formatting operation with comprehensive error handling
   * 
   * @returns Promise resolving to QualityAssuranceResult
   * @private
   */
  private static async executeFormatting(): Promise<QualityAssuranceResult> {
    const config: QualityAssuranceConfig = {
      timeout: OPERATION_TIMEOUTS.FORMATTING,
      command: QA_COMMANDS.FORMATTING,
      errorLogFile: ERROR_LOG_FILES.FORMATTING,
    };
    
    return this.executeQualityAssuranceOperation('formatting', config);
  }
  
  // ==========================================================================
  // Core Execution Logic
  // ==========================================================================
  
  /**
   * Generic quality assurance operation executor
   * 
   * @param operationType - Type of operation being executed
   * @param config - Configuration for the operation
   * @returns Promise resolving to QualityAssuranceResult
   * @private
   */
  private static async executeQualityAssuranceOperation(
    operationType: QualityOperationType,
    config: QualityAssuranceConfig
  ): Promise<QualityAssuranceResult> {
    this.validateOperationConfig(config);
    
    const context: ExecutionContext = {
      operationType,
      startTime: Date.now(),
    };
    
    try {
      Logger.debug(`Starting ${operationType} operation`, { command: config.command });
      
      const result = await executeShellCommand(config.command, { 
        timeout: config.timeout 
      });
      
      const duration = Date.now() - context.startTime;
      
      if (result.exitCode === 0) {
        Logger.debug(`${operationType} completed successfully`, { 
          duration,
          exitCode: result.exitCode 
        });
        return { success: true, duration, exitCode: result.exitCode };
      }
      
      // Operation failed - collect error information
      const errorMessage = this.buildDetailedErrorMessage(
        operationType,
        result.exitCode,
        result.stderr,
        result.stdout
      );
      
      await this.writeErrorLog(config.errorLogFile, errorMessage);
      
      Logger.error(`${operationType} failed`, {
        exitCode: result.exitCode,
        duration,
        hasStderr: Boolean(result.stderr),
        hasStdout: Boolean(result.stdout),
      });
      
      return {
        success: false,
        errorMessage,
        duration,
        exitCode: result.exitCode,
      };
      
    } catch (error) {
      const duration = Date.now() - context.startTime;
      const errorMessage = this.extractErrorMessage(error);
      
      await this.writeErrorLog(config.errorLogFile, errorMessage);
      
      Logger.error(`${operationType} threw exception`, {
        error: errorMessage,
        duration,
      });
      
      return {
        success: false,
        errorMessage,
        duration,
      };
    }
  }
  
  // ==========================================================================
  // Error Handling and Formatting
  // ==========================================================================
  
  /**
   * Builds detailed error message from command execution results
   * 
   * @param operationType - Type of operation that failed
   * @param exitCode - Exit code from the command
   * @param stderr - Standard error output
   * @param stdout - Standard output
   * @returns Formatted error message
   * @private
   */
  private static buildDetailedErrorMessage(
    operationType: QualityOperationType,
    exitCode: number,
    stderr: string,
    stdout: string
  ): string {
    const parts: string[] = [`${operationType} failed with exit code ${exitCode}`];
    
    if (stderr?.trim()) {
      parts.push(`stderr: ${stderr.trim()}`);
    }
    
    if (stdout?.trim()) {
      parts.push(`stdout: ${stdout.trim()}`);
    }
    
    if (!stderr?.trim() && !stdout?.trim()) {
      parts.push(MESSAGES.NO_OUTPUT);
    }
    
    return parts.join('\n');
  }
  
  /**
   * Formats error message with consistent structure
   * 
   * @param baseMessage - Base error message
   * @param additionalInfo - Additional error information
   * @returns Formatted error message
   * @private
   */
  private static formatErrorMessage(
    baseMessage: string, 
    additionalInfo?: string
  ): string {
    if (!additionalInfo) {
      return baseMessage;
    }
    
    return `${baseMessage}\n${additionalInfo}`;
  }
  
  /**
   * Safely extracts error message from unknown error type
   * 
   * @param error - Error object of unknown type
   * @returns String representation of the error
   * @private
   */
  private static extractErrorMessage(error: unknown): string {
    if (error instanceof Error) {
      return error.message;
    }
    
    if (typeof error === 'string') {
      return error;
    }
    
    if (error && typeof error === 'object' && 'toString' in error) {
      return String(error);
    }
    
    return 'Unknown error occurred';
  }
  
  /**
   * Writes error information to log file for debugging
   * 
   * @param filename - Name of the error log file
   * @param errorMessage - Error message to write
   * @returns Promise that resolves when writing is complete
   * @private
   */
  private static async writeErrorLog(
    filename: string, 
    errorMessage: string
  ): Promise<void> {
    try {
      const file = Bun.file(filename);
      await file.write(errorMessage);
    } catch (writeError) {
      Logger.error('Failed to write error log', {
        filename,
        writeError: this.extractErrorMessage(writeError),
        originalError: errorMessage,
      });
    }
  }
  
  // ==========================================================================
  // Validation
  // ==========================================================================
  
  /**
   * Validates quality assurance operation configuration
   * 
   * @param config - Configuration to validate
   * @throws Error if configuration is invalid
   * @private
   */
  private static validateOperationConfig(config: QualityAssuranceConfig): void {
    if (!config.command?.trim()) {
      throw new Error('Operation command cannot be empty');
    }
    
    if (config.timeout <= 0) {
      throw new Error('Operation timeout must be positive');
    }
    
    if (!config.errorLogFile?.trim()) {
      throw new Error('Error log file name cannot be empty');
    }
  }
}

// ============================================================================
// Module Entry Point
// ============================================================================

if (import.meta.main) {
  const result = await StopHook.execute();
  process.exit(result.exit_code);
}
